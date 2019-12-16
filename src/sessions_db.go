package main

import (
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	uuid "github.com/hashicorp/go-uuid"
	"go.mongodb.org/mongo-driver/bson"
)

type sessionEvent struct {
	When int32  `json:"when"`
	Data string `json:"data"`
}

func (dbobj dbcon) createSessionRecord(userTOKEN string, expiration string, data []byte) (string, error) {
	var endtime int32 = 0
	var err error
	if len(expiration) > 0 {
		endtime, err = parseExpiration(expiration)
		if err != nil {
			return "", err
		}
	}
	encodedStr, err := dbobj.userEncrypt(userTOKEN, data)
	if err != nil {
		return "", err
	}
	sessionUUID, err := uuid.GenerateUUID()
	if err != nil {
		return "", err
	}
	bdoc := bson.M{}
	now := int32(time.Now().Unix())
	bdoc["token"] = userTOKEN
	bdoc["session"] = sessionUUID
	bdoc["endtime"] = endtime
	bdoc["when"] = now
	bdoc["data"] = encodedStr
	_, err = dbobj.createRecord(TblName.Sessions, bdoc)
	if err != nil {
		return "", err
	}
	return sessionUUID, nil
}

func (dbobj dbcon) getUserSession(sessionUUID string) (int32, []byte, string, error) {
	record, err := dbobj.getRecord(TblName.Sessions, "session", sessionUUID)
	if err != nil {
		return 0, nil, "", err
	}
	if record == nil {
		return 0, nil, "", errors.New("not found")
	}
	// check expiration
	now := int32(time.Now().Unix())
	if now > record["endtime"].(int32) {
		return 0, nil, "", errors.New("session expired")
	}
	when := record["when"].(int32)
	userTOKEN := record["token"].(string)
	encData0 := record["data"].(string)
	decrypted, err := dbobj.userDecrypt(userTOKEN, encData0)
	if err != nil {
		return 0, nil, "", err
	}
	return when, decrypted, userTOKEN, err
}

func (dbobj dbcon) getUserSessionByToken(userTOKEN string) ([]string, int64, error) {

	userBson, err := dbobj.lookupUserRecord(userTOKEN)
	if userBson == nil || err != nil {
		// not found
		return nil, 0, err
	}
	userKey := userBson["key"].(string)
	recordKey, err := base64.StdEncoding.DecodeString(userKey)
	if err != nil {
		return nil, 0, err
	}

	count, err := dbobj.countRecords(TblName.Sessions, "token", userTOKEN)
	if err != nil {
		return nil, 0, err
	}

	records, err := dbobj.getList(TblName.Sessions, "token", userTOKEN, 0, 0)
	if err != nil {
		return nil, 0, err
	}

	var results []string
	for _, element := range records {
		when := element["when"].(int32)
		encData0 := element["data"].(string)
		encData, _ := base64.StdEncoding.DecodeString(encData0)
		decrypted, _ := decrypt(dbobj.masterKey, recordKey, encData)
		sEvent := fmt.Sprintf(`{"when":%d,"data":%s}`, when, string(decrypted))
		results = append(results, sEvent)
	}

	return results, count, err
}
