package main

import (
    "fmt"
    "net/http"
    "reflect"

    "github.com/julienschmidt/httprouter"
    //"go.mongodb.org/mongo-driver/bson"
)

func (e mainEnv) createLegalBasis(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
    brief := ps.ByName("brief")
    authResult := e.enforceAdmin(w, r)
    if authResult == "" {
        return
    }
    brief = normalizeBrief(brief)
    if isValidBrief(brief) == false {
        returnError(w, r, "bad brief format", 405, nil, nil)
        return
    }
    records, err := getJSONPostData(r)
    if err != nil {
        returnError(w, r, "failed to decode request body", 405, err, nil)
        return
    }
    status := "active";
    module := getStringValue(records, "module")
    fulldesc := getStringValue(records, "fulldesc")
    newbrief := getStringValue(records, "newbrief")
    shortdesc := getStringValue(records, "shortdesc")
    basistype := getStringValue(records, "basistype")
    requiredmsg := getStringValue(records, "requiredmsg")
    usercontrol := false
    requiredflag := false
    if value, ok := records["status"]; ok {
        if reflect.TypeOf(value) == reflect.TypeOf("string") {
            if value.(string) == "disabled" {
                status = value.(string)
            }
        }
    }
    if value, ok := records["usercontrol"]; ok {
        if reflect.TypeOf(value).Kind() == reflect.Bool {
            usercontrol = value.(bool)
        }
    }
    if value, ok := records["requiredflag"]; ok {
        if reflect.TypeOf(value).Kind() == reflect.Bool {
            requiredflag = value.(bool)
        }
    }
    e.db.createLegalBasis(brief, newbrief, module, shortdesc, fulldesc, basistype, requiredmsg, status, usercontrol, requiredflag)
    /*
    notifyURL := e.conf.Notification.NotificationURL
    if newStatus == true && len(notifyURL) > 0 {
        // change notificate on new record or if status change
        if len(userTOKEN) > 0 {
            notifyConsentChange(notifyURL, brief, status, "token", userTOKEN)
        } else {
            notifyConsentChange(notifyURL, brief, status, mode, address)
        }
    }
    */
    w.Header().Set("Content-Type", "application/json; charset=utf-8")
    w.WriteHeader(200)
    w.Write([]byte(`{"status":"ok"}`))
}

func (e mainEnv) deleteLegalBasis(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
    brief := ps.ByName("brief")
    authResult := e.enforceAdmin(w, r)
    if authResult == "" {
        return
    }
    brief = normalizeBrief(brief)
    if isValidBrief(brief) == false {
        returnError(w, r, "bad brief format", 405, nil, nil)
        return
    }
    e.db.unlinkProcessingActivityBrief(brief)
    e.db.deleteLegalBasis(brief);
    w.Header().Set("Content-Type", "application/json; charset=utf-8")
    w.WriteHeader(200)
    w.Write([]byte(`{"status":"ok"}`))
}

func (e mainEnv) listLegalBasisRecords(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
    authResult := e.enforceAdmin(w, r)
    if authResult == "" {
        return
    }
    resultJSON, numRecords, err := e.db.getLegalBasisRecords()
    if err != nil {
        returnError(w, r, "internal error", 405, err, nil)
        return
    }
    fmt.Printf("Total count of rows: %d\n", numRecords)
    w.Header().Set("Content-Type", "application/json; charset=utf-8")
    w.WriteHeader(200)
    str := fmt.Sprintf(`{"status":"ok","total":%d,"rows":%s}`, numRecords, resultJSON)
    w.Write([]byte(str))
}
