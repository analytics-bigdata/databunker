package storage

// https://stackoverflow.com/questions/21986780/is-it-possible-to-retrieve-a-column-value-by-name-using-golang-database-sql

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3" // load sqlite3 here
	"github.com/schollz/sqlite3dump"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	knownApps []string
)

// Tbl is used to store table id
type Tbl int

// listTbls used to store list of tables
type listTbls struct {
	Users         Tbl
	Audit         Tbl
	Xtokens       Tbl
	Sessions      Tbl
	Requests      Tbl
	Legalbasis    Tbl
	Agreements    Tbl
	Sharedrecords Tbl
	Processingactivities Tbl
}

// TblName is enum of tables
var TblName = &listTbls{
	Users:         0,
	Audit:         1,
	Xtokens:       2,
	Sessions:      3,
	Requests:      4,
	Legalbasis:    5,
	Agreements:    6,
	Sharedrecords: 7,
	Processingactivities: 8,
}

// DBStorage struct is used to store database object
type DBStorage struct {
	db *sql.DB
}

// DBExists function checks if database exists
func DBExists(filepath *string) bool {
	dbfile := "./databunker.db"
	if filepath != nil {
		if len(*filepath) > 0 {
			dbfile = *filepath
		}
	}
	if len(dbfile) >= 3 && dbfile[len(dbfile)-3:] != ".db" {
		dbfile = dbfile + ".db"
	}
	if _, err := os.Stat(dbfile); os.IsNotExist(err) {
		return false
	}
	return true
}

// CreateTestDB creates a test db
func CreateTestDB() string {
	testDBFile := "/tmp/test-sqlite.db"
	os.Remove(testDBFile)
	return testDBFile
}
// OpenDB function opens the database
func OpenDB(filepath *string) (DBStorage, error) {
	dbfile := "./databunker.db"
	if filepath != nil {
		if len(*filepath) > 0 {
			dbfile = *filepath
		}
	}
	if len(dbfile) >= 3 && dbfile[len(dbfile)-3:] != ".db" {
		dbfile = dbfile + ".db"
	}
	fmt.Printf("Databunker db file is: %s\n", dbfile)
	// collect list of all tables
	/*
		if _, err := os.Stat(dbfile); !os.IsNotExist(err) {
			db2, err := ql.OpenFile(dbfile, &ql.Options{FileFormat: 2})
			if err != nil {
				return dbobj, err
			}
			dbinfo, err := db2.Info()
			for _, v := range dbinfo.Tables {
				knownApps = append(knownApps, v.Name)
			}
			db2.Close()
		}
	*/

	//ql.RegisterDriver2()
	//db, err := sql.Open("ql2", dbfile)
	db, err := sql.Open("sqlite3", "file:"+dbfile+"?_journal_mode=WAL")
	if err != nil {
		log.Fatalf("Failed to open databunker.db file: %s", err)
	}
	err = db.Ping()
	if err != nil {
		log.Fatalf("Error on opening database connection: %s", err.Error())
	}
	_, err = db.Exec("vacuum")
	if err != nil {
		log.Fatalf("Error on vacuum database command")
	}
	dbobj := DBStorage{db}

	// load all table names
	q := "select name from sqlite_master where type ='table'"
	tx, err := dbobj.db.Begin()
	if err != nil {
		return dbobj, err
	}
	defer tx.Rollback()
	rows, err := tx.Query(q)
	for rows.Next() {
		t := ""
		rows.Scan(&t)
		knownApps = append(knownApps, t)
	}
	tx.Commit()
	fmt.Printf("tables: %s\n", knownApps)
	return dbobj, nil
}

// InitDB function creates tables and indexes
func InitDB(filepath *string) (DBStorage, error){
	dbfile := "./databunker.db"
	if filepath != nil {
		if len(*filepath) > 0 {
			dbfile = *filepath
		}
	}
	if len(dbfile) >= 3 && dbfile[len(dbfile)-3:] != ".db" {
		dbfile = dbfile + ".db"
	}
	fmt.Printf("Init Databunker db file is: %s\n", dbfile)
	db, err := sql.Open("sqlite3", "file:"+dbfile+"?_journal_mode=WAL")
	if err != nil {
		log.Fatalf("Failed to open databunker.db file: %s", err)
	}
	err = db.Ping()
	if err != nil {
		log.Fatalf("Error on opening database connection: %s", err.Error())
	}
	dbobj := DBStorage{db}
	
	initUsers(dbobj.db)
	initXTokens(dbobj.db)
	initAudit(dbobj.db)
	initSessions(dbobj.db)
	initRequests(dbobj.db)
	initSharedRecords(dbobj.db)
	initProcessingactivities(dbobj.db)
	initLegalbasis(dbobj.db)
	initAgreements(dbobj.db)
	return dbobj, nil
}


func (dbobj DBStorage) Ping() error {
        return dbobj.db.Ping()
}

// CloseDB function closes the open database
func (dbobj DBStorage) CloseDB() {
	dbobj.db.Close()
}

// BackupDB function backups existing databsae and prints database structure to http.ResponseWriter
func (dbobj DBStorage) BackupDB(w http.ResponseWriter) {
	err := sqlite3dump.DumpDB(dbobj.db, w)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Printf("error in backup: %s", err)
	}
}

// InitUserApps initialises list of databases.
func (dbobj DBStorage) InitUserApps() error {
	return nil
}

func escapeName(name string) string {
	if name == "when" {
		name = "`when`"
	}
	return name
}

func decodeFieldsValues(data interface{}) (string, []interface{}) {
	fields := ""
	values := make([]interface{}, 0)

	switch t := data.(type) {
	case primitive.M:
		//fmt.Println("decodeFieldsValues format is: primitive.M")
		for idx, val := range data.(primitive.M) {
			if len(fields) == 0 {
				fields = escapeName(idx)
			} else {
				fields = fields + "," + escapeName(idx)
			}
			values = append(values, val)
		}
	case *primitive.M:
		//fmt.Println("decodeFieldsValues format is: *primitive.M")
		for idx, val := range *data.(*primitive.M) {
			if len(fields) == 0 {
				fields = escapeName(idx)
			} else {
				fields = fields + "," + escapeName(idx)
			}
			values = append(values, val)
		}
	case map[string]interface{}:
		//fmt.Println("decodeFieldsValues format is: map[string]interface{}")
		for idx, val := range data.(map[string]interface{}) {
			if len(fields) == 0 {
				fields = escapeName(idx)
			} else {
				fields = fields + "," + escapeName(idx)
			}
			values = append(values, val)
		}
	default:
		fmt.Printf("XXXXXX wrong type: %T\n", t)
	}
	return fields, values
}

func decodeForCleanup(data interface{}) string {
	fields := ""

	switch t := data.(type) {
	case primitive.M:
		for idx := range data.(primitive.M) {
			if len(fields) == 0 {
				fields = escapeName(idx) + "=null"
			} else {
				fields = fields + "," + escapeName(idx) + "=null"
			}
		}
		return fields
	case map[string]interface{}:
		for idx := range data.(map[string]interface{}) {
			if len(fields) == 0 {
				fields = escapeName(idx) + "=null"
			} else {
				fields = fields + "," + escapeName(idx) + "=null"
			}
		}
	default:
		fmt.Printf("decodeForCleanup: wrong type: %s\n", t)
	}

	return fields
}

func decodeForUpdate(bdoc *bson.M, bdel *bson.M) (string, []interface{}) {
	values := make([]interface{}, 0)
	fields := ""

	if bdoc != nil {
		/*
			switch t := *bdoc.(type) {
			default:
				fmt.Printf("Type is %T\n", t)
			}
		*/

		for idx, val := range *bdoc {
			values = append(values, val)
			if len(fields) == 0 {
				fields = escapeName(idx) + "=$1"
			} else {
				fields = fields + "," + escapeName(idx) + "=$" + (strconv.Itoa(len(values)))
			}
		}
	}

	if bdel != nil {
		for idx := range *bdel {
			if len(fields) == 0 {
				fields = escapeName(idx) + "=null"
			} else {
				fields = fields + "," + escapeName(idx) + "=null"
			}
		}
	}
	return fields, values
}

func getTable(t Tbl) string {
	switch t {
	case TblName.Users:
		return "users"
	case TblName.Audit:
		return "audit"
	case TblName.Xtokens:
		return "xtokens"
	case TblName.Sessions:
		return "sessions"
	case TblName.Requests:
		return "requests"
	case TblName.Legalbasis:
		return "legalbasis"
	case TblName.Agreements:
		return "agreements"
	case TblName.Sharedrecords:
		return "sharedrecords"
	case TblName.Processingactivities:
		return "processingactivities"
	}
	return "users"
}

// CreateRecordInTable creates new record
func (dbobj DBStorage) CreateRecordInTable(tbl string, data interface{}) (int, error) {
	fields, values := decodeFieldsValues(data)
	valuesInQ := "$1"
	for idx := range values {
		if idx > 0 {
			valuesInQ = valuesInQ + ",$" + (strconv.Itoa(idx + 1))
		}
	}
	q := "insert into " + tbl + " (" + fields + ") values (" + valuesInQ + ")"
	fmt.Printf("q: %s\n", q)
	//fmt.Printf("values: %s\n", values...)
	tx, err := dbobj.db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()
	_, err = tx.Exec(q, values...)
	if err != nil {
		return 0, err
	}
	if err = tx.Commit(); err != nil {
		return 0, err
	}
	return 1, nil
}

// CreateRecord creates new record
func (dbobj DBStorage) CreateRecord(t Tbl, data interface{}) (int, error) {
	//if reflect.TypeOf(value) == reflect.TypeOf("string")
	tbl := getTable(t)
	return dbobj.CreateRecordInTable(tbl, data)
}

// CountRecords returns number of records in table
func (dbobj DBStorage) CountRecords0(t Tbl) (int64, error) {
	tbl := getTable(t)
	q := "select count(*) from " + tbl
	fmt.Printf("q: %s\n", q)

	tx, err := dbobj.db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()
	row := tx.QueryRow(q)
	// Columns
	var count int
	err = row.Scan(&count)
	if err != nil {
		return 0, err
	}
	if err = tx.Commit(); err != nil {
		return 0, err
	}
	return int64(count), nil
}

// CountRecords returns number of records that match filter
func (dbobj DBStorage) CountRecords(t Tbl, keyName string, keyValue string) (int64, error) {
        tbl := getTable(t)
        q := "select count(*) from " + tbl + " WHERE " + escapeName(keyName) + "=$1"
        fmt.Printf("q: %s\n", q)

        tx, err := dbobj.db.Begin()
        if err != nil {
                return 0, err
        }
        defer tx.Rollback()
        row := tx.QueryRow(q, keyValue)
        // Columns
        var count int
        err = row.Scan(&count)
        if err != nil {
                return 0, err
        }
        if err = tx.Commit(); err != nil {
                return 0, err
        }
        return int64(count), nil
}

// UpdateRecord updates database record
func (dbobj DBStorage) UpdateRecord(t Tbl, keyName string, keyValue string, bdoc *bson.M) (int64, error) {
	table := getTable(t)
	filter := escapeName(keyName) + "=\"" + keyValue + "\""
	return dbobj.updateRecordInTableDo(table, filter, bdoc, nil)
}

// UpdateRecordInTable updates database record
func (dbobj DBStorage) UpdateRecordInTable(table string, keyName string, keyValue string, bdoc *bson.M) (int64, error) {
	filter := escapeName(keyName) + "=\"" + keyValue + "\""
	return dbobj.updateRecordInTableDo(table, filter, bdoc, nil)
}

// UpdateRecord2 updates database record
func (dbobj DBStorage) UpdateRecord2(t Tbl, keyName string, keyValue string,
	keyName2 string, keyValue2 string, bdoc *bson.M, bdel *bson.M) (int64, error) {
	table := getTable(t)
	filter := escapeName(keyName) + "=\"" + keyValue + "\" AND " +
		escapeName(keyName2) + "=\"" + keyValue2 + "\""
	return dbobj.updateRecordInTableDo(table, filter, bdoc, bdel)
}

// UpdateRecordInTable2 updates database record
func (dbobj DBStorage) UpdateRecordInTable2(table string, keyName string,
	keyValue string, keyName2 string, keyValue2 string, bdoc *bson.M, bdel *bson.M) (int64, error) {
	filter := escapeName(keyName) + "=\"" + keyValue + "\" AND " +
		escapeName(keyName2) + "=\"" + keyValue2 + "\""
	return dbobj.updateRecordInTableDo(table, filter, bdoc, bdel)
}

func (dbobj DBStorage) updateRecordInTableDo(table string, filter string, bdoc *bson.M, bdel *bson.M) (int64, error) {
	op, values := decodeForUpdate(bdoc, bdel)
	q := "update " + table + " SET " + op + " WHERE " + filter
	fmt.Printf("q: %s\n", q)

	tx, err := dbobj.db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()
	result, err := tx.Exec(q, values...)
	if err != nil {
		return 0, err
	}
	if err = tx.Commit(); err != nil {
		return 0, err
	}
	num, err := result.RowsAffected()
	return num, err
}

// Lookup record by multiple fields
func (dbobj DBStorage) LookupRecord(t Tbl, row bson.M) (bson.M, error) {
	table := getTable(t)
	q := "select * from " + table + " WHERE "
	num := 1
	values := make([]interface{}, 0)
	for keyName, keyValue := range row {
		q = q + escapeName(keyName) + "=$" + strconv.FormatInt(int64(num), 10)
		if num < len(row) {
			q = q + " AND "
		}
		values = append(values, keyValue)
		num = num + 1
	}
	return dbobj.getRecordInTableDo(q, values)
}

// GetRecord returns specific record from database
func (dbobj DBStorage) GetRecord(t Tbl, keyName string, keyValue string) (bson.M, error) {
	table := getTable(t)
	q := "select * from " + table + " WHERE " + escapeName(keyName) + "=$1"
	values := make([]interface{}, 0)
	values = append(values, keyValue)
	return dbobj.getRecordInTableDo(q, values)
}

// GetRecordInTable returns specific record from database
func (dbobj DBStorage) GetRecordInTable(table string, keyName string, keyValue string) (bson.M, error) {
	q := "select * from " + table + " WHERE " + escapeName(keyName) + "=$1"
	values := make([]interface{}, 0)
	values = append(values, keyValue)
	return dbobj.getRecordInTableDo(q, values)
}

// GetRecord2  returns specific record from database
func (dbobj DBStorage) GetRecord2(t Tbl, keyName string, keyValue string,
	keyName2 string, keyValue2 string) (bson.M, error) {
	table := getTable(t)
	q := "select * from " + table + " WHERE " + escapeName(keyName) + "=$1 AND " +
		escapeName(keyName2) + "=$2"
	values := make([]interface{}, 0)
	values = append(values, keyValue)
	values = append(values, keyValue2)
	return dbobj.getRecordInTableDo(q, values)
}

func (dbobj DBStorage) getRecordInTableDo(q string, values []interface{}) (bson.M, error) {
	fmt.Printf("query: %s\n", q)

	tx, err := dbobj.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	rows, err := tx.Query(q, values...)
	if err == sql.ErrNoRows {
		fmt.Println("nothing found")
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	defer rows.Close()

	columnNames, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	//fmt.Printf("names: %s\n", columnNames)
	if err := rows.Err(); err != nil {
		return nil, err
	}
	//pointers := make([]interface{}, len(columnNames))
	recBson := bson.M{}
	rows.Next()
	columnPointers := make([]interface{}, len(columnNames))
	//for i, _ := range columnNames {
	//		columnPointers[i] = new(interface{})
	//}
	columns := make([]interface{}, len(columnNames))
	for idx := range columns {
		columnPointers[idx] = &columns[idx]
	}
	err = rows.Scan(columnPointers...)
	if err == sql.ErrNoRows {
		fmt.Println("returning empty result")
		return nil, nil
	}
	if err != nil {
		if strings.Contains(err.Error(), "Rows are closed") {
			fmt.Println("returning empty result")
			return nil, nil
		}
		return nil, err
	}
	for i, colName := range columnNames {
		switch t := columns[i].(type) {
		case string:
			recBson[colName] = columns[i]
		case []uint8:
			recBson[colName] = string(columns[i].([]uint8))
		case int64:
			recBson[colName] = int32(columns[i].(int64))
		case int32:
			recBson[colName] = int32(columns[i].(int32))
		case bool:
			recBson[colName] = columns[i].(bool)
		case nil:
			//fmt.Printf("is nil, not interesting\n")
		default:
			fmt.Printf("field: %s - %s, unknown: %s - %T\n", colName, columns[i], t, t)
		}
	}
	err = rows.Close()
	if err == sql.ErrNoRows {
		fmt.Println("nothing found2")
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	if len(recBson) == 0 {
		fmt.Println("no result!!!")
		return nil, nil
	}
	tx.Commit()
	return recBson, nil
}

// DeleteRecord deletes record in database
func (dbobj DBStorage) DeleteRecord(t Tbl, keyName string, keyValue string) (int64, error) {
	tbl := getTable(t)
	return dbobj.DeleteRecordInTable(tbl, keyName, keyValue)
}

// DeleteRecordInTable deletes record in database
func (dbobj DBStorage) DeleteRecordInTable(table string, keyName string, keyValue string) (int64, error) {
	q := "delete from " + table + " WHERE " + escapeName(keyName) + "=$1"
	fmt.Printf("q: %s\n", q)

	tx, err := dbobj.db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()
	result, err := tx.Exec(q, keyValue)
	if err != nil {
		return 0, err
	}
	if err = tx.Commit(); err != nil {
		return 0, err
	}
	num, err := result.RowsAffected()
	return num, err
}

// DeleteRecord2 deletes record in database
func (dbobj DBStorage) DeleteRecord2(t Tbl, keyName string, keyValue string, keyName2 string, keyValue2 string) (int64, error) {
	tbl := getTable(t)
	return dbobj.deleteRecordInTable2(tbl, keyName, keyValue, keyName2, keyValue2)
}

func (dbobj DBStorage) deleteRecordInTable2(table string, keyName string, keyValue string, keyName2 string, keyValue2 string) (int64, error) {
	q := "delete from " + table + " WHERE " + escapeName(keyName) + "=$1 AND " +
		escapeName(keyName2) + "=$2"
	fmt.Printf("q: %s\n", q)

	tx, err := dbobj.db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()
	result, err := tx.Exec(q, keyValue, keyValue2)
	if err != nil {
		return 0, err
	}
	if err = tx.Commit(); err != nil {
		return 0, err
	}
	num, err := result.RowsAffected()
	return num, err
}

/*
func (dbobj DBStorage) deleteDuplicate2(t Tbl, keyName string, keyValue string, keyName2 string, keyValue2 string) (int64, error) {
	tbl := getTable(t)
	return dbobj.deleteDuplicateInTable2(tbl, keyName, keyValue, keyName2, keyValue2)
}
*/

/*
func (dbobj DBStorage) deleteDuplicateInTable2(table string, keyName string, keyValue string, keyName2 string, keyValue2 string) (int64, error) {
	q := "delete from " + table + " where " + escapeName(keyName) + "=$1 AND " +
		escapeName(keyName2) + "=$2 AND rowid not in " +
		"(select min(rowid) from " + table + " where " + escapeName(keyName) + "=$3 AND " +
		escapeName(keyName2) + "=$4)"
	fmt.Printf("q: %s\n", q)

	tx, err := dbobj.db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()
	result, err := tx.Exec(q, keyValue, keyValue2, keyValue, keyValue2)
	if err != nil {
		return 0, err
	}
	if err = tx.Commit(); err != nil {
		return 0, err
	}
	num, err := result.RowsAffected()
	return num, err
}
*/

// DeleteExpired0 deletes expired records in database
func (dbobj DBStorage) DeleteExpired0(t Tbl, expt int32) (int64, error) {
	table := getTable(t)
	now := int32(time.Now().Unix())
	q := fmt.Sprintf("delete from %s WHERE `when`>0 AND `when`<%d", table, now-expt)
	fmt.Printf("q: %s\n", q)
	tx, err := dbobj.db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()
	result, err := tx.Exec(q)
	if err != nil {
		return 0, err
	}
	if err = tx.Commit(); err != nil {
		return 0, err
	}
	num, err := result.RowsAffected()
	// vacuum database
	dbobj.db.Exec("vacuum")
	return num, err
}

// DeleteExpired deletes expired records in database
func (dbobj DBStorage) DeleteExpired(t Tbl, keyName string, keyValue string) (int64, error) {
	table := getTable(t)
	q := "delete from " + table + " WHERE endtime>0 AND endtime<$1 AND " + escapeName(keyName) + "=$2"
	fmt.Printf("q: %s\n", q)

	tx, err := dbobj.db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()
	now := int32(time.Now().Unix())
	result, err := tx.Exec(q, now, keyValue)
	if err != nil {
		return 0, err
	}
	if err = tx.Commit(); err != nil {
		return 0, err
	}
	num, err := result.RowsAffected()
	return num, err
}

// CleanupRecord nullifies specific feilds in records in database
func (dbobj DBStorage) CleanupRecord(t Tbl, keyName string, keyValue string, data interface{}) (int64, error) {
	tbl := getTable(t)
	cleanup := decodeForCleanup(data)
	q := "update " + tbl + " SET " + cleanup + " WHERE " + escapeName(keyName) + "=$1"
	fmt.Printf("q: %s\n", q)

	tx, err := dbobj.db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()
	result, err := tx.Exec(q, keyValue)
	if err != nil {
		return 0, err
	}
	if err = tx.Commit(); err != nil {
		return 0, err
	}
	num, err := result.RowsAffected()
	return num, err
}

// GetExpiring get records that are expiring
func (dbobj DBStorage) GetExpiring(t Tbl, keyName string, keyValue string) ([]bson.M, error) {
	table := getTable(t)
	now := int32(time.Now().Unix())
	q := fmt.Sprintf("select * from %s WHERE endtime>0 AND endtime<%d AND %s=$1", table, now, escapeName(keyName))
	fmt.Printf("q: %s\n", q)
	values := make([]interface{}, 0)
	values = append(values, keyValue)
	return dbobj.getListDo(q, values)
}

// GetUniqueList returns a unique list of values from specific column in database
func (dbobj DBStorage) GetUniqueList(t Tbl, keyName string) ([]bson.M, error) {
	table := getTable(t)
	keyName = escapeName(keyName)
	q := "select distinct " + keyName + " from " + table + " ORDER BY " + keyName
	fmt.Printf("q: %s\n", q)
	values := make([]interface{}, 0)
	return dbobj.getListDo(q, values)
}

// GetList is used to return list of rows. It can be used to return values using pager.
func (dbobj DBStorage) GetList0(t Tbl, start int32, limit int32, orderField string) ([]bson.M, error) {
	table := getTable(t)
	if limit > 100 {
		limit = 100
	}

	q := "select * from " + table
	if len(orderField) > 0 {
		q = q + " ORDER BY " + escapeName(orderField) + " DESC"
	}
	if start > 0 {
		q = q + " LIMIT " + strconv.FormatInt(int64(limit), 10) +
			" OFFSET " + strconv.FormatInt(int64(start), 10)
	} else if limit > 0 {
		q = q + " LIMIT " + strconv.FormatInt(int64(limit), 10)
	}
	fmt.Printf("q: %s\n", q)
	values := make([]interface{}, 0)
	return dbobj.getListDo(q, values)
}

// GetList is used to return list of rows. It can be used to return values using pager.
func (dbobj DBStorage) GetList(t Tbl, keyName string, keyValue string, start int32, limit int32, orderField string) ([]bson.M, error) {
	table := getTable(t)
	if limit > 100 {
		limit = 100
	}

	q := "select * from " + table + " WHERE " + escapeName(keyName) + "=$1"
	if len(orderField) > 0 {
		q = q + " ORDER BY " + escapeName(orderField) + " DESC"
	}
	if start > 0 {
		q = q + " LIMIT " + strconv.FormatInt(int64(limit), 10) +
			" OFFSET " + strconv.FormatInt(int64(start), 10)
	} else if limit > 0 {
		q = q + " LIMIT " + strconv.FormatInt(int64(limit), 10)
	}
	fmt.Printf("q: %s\n", q)
	values := make([]interface{}, 0)
	values = append(values, keyValue)
	return dbobj.getListDo(q, values)
}

func (dbobj DBStorage) getListDo(q string, values []interface{}) ([]bson.M, error) {
	tx, err := dbobj.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	rows, err := tx.Query(q, values...)
	if err == sql.ErrNoRows {
		fmt.Println("nothing found")
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	defer rows.Close()

	columnNames, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	//fmt.Printf("names: %s\n", columnNames)
	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}
	var results []bson.M
	//pointers := make([]interface{}, len(columnNames))
	//rows.Next()
	for rows.Next() {
		recBson := bson.M{}
		//fmt.Println("parsing result line")
		columnPointers := make([]interface{}, len(columnNames))
		//for i, _ := range columnNames {
		//		columnPointers[i] = new(interface{})
		//}
		columns := make([]interface{}, len(columnNames))
		for idx := range columns {
			columnPointers[idx] = &columns[idx]
		}

		err = rows.Scan(columnPointers...)
		if err == sql.ErrNoRows {
			fmt.Println("nothing found")
			return nil, nil
		}
		if err != nil {
			fmt.Printf("nothing found: %s\n", err)
			return nil, nil
		}
		for i, colName := range columnNames {
			switch t := columns[i].(type) {
			case string:
				recBson[colName] = columns[i]
			case []uint8:
				recBson[colName] = string(columns[i].([]uint8))
			case int64:
				recBson[colName] = int32(columns[i].(int64))
			case bool:
				recBson[colName] = columns[i].(bool)
			case nil:
				//fmt.Printf("is nil, not interesting\n")
			default:
				fmt.Printf("field: %s - %s, unknown: %s - %T\n", colName, columns[i], t, t)
			}
		}
		results = append(results, recBson)
	}
	err = rows.Close()
	if err == sql.ErrNoRows {
		fmt.Println("nothing found2")
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		fmt.Println("no result!!!")
		return nil, nil
	}
	tx.Commit()
	return results, nil
}

// GetAllTables returns all tables that exists in database
func (dbobj DBStorage) GetAllTables() ([]string, error) {
	return knownApps, nil
}

// ValidateNewApp function check if app name can be part of the table name
func (dbobj DBStorage) ValidateNewApp(appName string) bool {
	if contains(knownApps, appName) == true {
		return true
	}
	if len(knownApps) >= 10 {
		return false
	}
	return true
}

func execQueries(db *sql.DB, queries []string) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for _, value := range queries {
		_, err = tx.Exec(value)
		if err != nil {
			return err
		}
	}
	if err = tx.Commit(); err != nil {
		return err
	}
	return nil
}

// IndexNewApp creates a new app table and creates indexes for it.
func (dbobj DBStorage) IndexNewApp(appName string) {
	if contains(knownApps, appName) == false {
		// it is a new app, create an index
		fmt.Printf("This is a new app, creating table & index for: %s\n", appName)
		queries := []string{"CREATE TABLE IF NOT EXISTS " + appName + ` (
			token STRING,
			md5 STRING,
			rofields STRING,
			data TEXT,
			status STRING,
			` + "`when` int);",
			"CREATE INDEX " + appName + "_token ON " + appName + " (token);"}
		err := execQueries(dbobj.db, queries)
		if err == nil {
			knownApps = append(knownApps, appName)
		}
	}
	return
}

func initUsers(db *sql.DB) error {
	queries := []string{`CREATE TABLE IF NOT EXISTS users (
			  token STRING,
			  key STRING,
			  md5 STRING,
			  loginidx STRING,
			  emailidx STRING,
			  phoneidx STRING,
			  rofields STRING,
			  tempcodeexp int,
			  tempcode int,
			  data TEXT
			);`,
		`CREATE INDEX users_token ON users (token);`,
		`CREATE INDEX users_login ON users (loginidx);`,
		`CREATE INDEX users_email ON users (emailidx);`,
		`CREATE INDEX users_phone ON users (phoneidx);`}
	return execQueries(db, queries)
}

func initXTokens(db *sql.DB) error {
	queries := []string{`CREATE TABLE IF NOT EXISTS xtokens (
				  xtoken STRING,
				  token STRING,
				  type STRING,
				  app STRING,
				  fields STRING,
				  endtime int
				);`,
		`CREATE UNIQUE INDEX xtokens_xtoken ON xtokens (xtoken);`,
		`CREATE INDEX xtokens_uniq ON xtokens (token, type);`}
	return execQueries(db, queries)
}

func initSharedRecords(db *sql.DB) error {
	queries := []string{`CREATE TABLE IF NOT EXISTS sharedrecords (
				  token STRING,
				  record STRING,
				  partner STRING,
				  session STRING,
				  app STRING,
				  fields STRING,
				  endtime int,
				  ` + "`when` int);",
		`CREATE INDEX sharedrecords_record ON sharedrecords (record);`}
	return execQueries(db, queries)
}

func initAudit(db *sql.DB) error {
	queries := []string{`CREATE TABLE IF NOT EXISTS audit (
				  atoken STRING,
				  identity STRING,
				  record STRING,
				  who STRING,
				  mode STRING,
				  app STRING,
				  title STRING,
				  status STRING,
				  msg STRING,
				  debug STRING,
				  before TEXT,
				  after TEXT,
				  ` + "`when` int);",
		`CREATE INDEX audit_atoken ON audit (atoken);`,
		`CREATE INDEX audit_record ON audit (record);`}
	return execQueries(db, queries)
}

func initRequests(db *sql.DB) error {
	queries := []string{`CREATE TABLE IF NOT EXISTS requests (
				  rtoken STRING,
				  token STRING,
				  app STRING,
				  brief STRING,
				  action STRING,
				  status STRING,
				  change STRING,
				  reason STRING,
				  creationtime int,
				  ` + "`when` int);",
		`CREATE INDEX requests_rtoken ON requests (rtoken);`,
		`CREATE INDEX requests_token ON requests (token);`,
		`CREATE INDEX requests_status ON requests (status);`}
	return execQueries(db, queries)
}

func initProcessingactivities(db *sql.DB) error {
	queries := []string{`CREATE TABLE IF NOT EXISTS processingactivities (
				  activity STRING,
				  title STRING,
				  script STRING,
				  fulldesc STRING,
				  legalbasis STRING,
				  applicableto STRING,
				  creationtime int);`,
		`CREATE INDEX processingactivities_activity ON processingactivities (activity);`}
	return execQueries(db, queries)
}

func initLegalbasis(db *sql.DB) error {
	queries := []string{`CREATE TABLE IF NOT EXISTS legalbasis (
				  brief STRING,
				  status STRING,
				  module STRING,
				  shortdesc STRING,
				  fulldesc STRING,
				  basistype STRING,
				  requiredmsg STRING,
				  usercontrol BOOLEAN,
				  requiredflag BOOLEAN,
				  creationtime int);`,
		`CREATE INDEX legalbasis_brief ON legalbasis (brief);`}
	return execQueries(db, queries)
}

func initAgreements(db *sql.DB) error {
	queries := []string{`CREATE TABLE IF NOT EXISTS agreements (
				  who STRING,
				  mode STRING,
				  token STRING,
				  brief STRING,
				  status STRING,
				  referencecode STRING,
				  lastmodifiedby STRING,
				  agreementmethod STRING,
				  creationtime int,
				  starttime int,
				  endtime int,
				  ` + "`when` int);",
		`CREATE INDEX agreements_token ON agreements (token);`,
		`CREATE INDEX agreements_brief ON agreements (brief);`}
	return execQueries(db, queries)
}

func initSessions(db *sql.DB) error {
	queries := []string{`CREATE TABLE IF NOT EXISTS sessions (
				  token STRING,
				  session STRING,
				  data TEXT,
				  endtime int,
				  ` + "`when` int);",
		`CREATE INDEX sessions_token ON sessions (token);`,
		`CREATE INDEX sessions_session ON sessions (session);`}
	return execQueries(db, queries)
}

func contains(slice []string, item string) bool {
	set := make(map[string]struct{}, len(slice))
	for _, s := range slice {
		set[s] = struct{}{}
	}

	_, ok := set[item]
	return ok
}
