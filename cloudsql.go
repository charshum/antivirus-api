// Copyright 2019 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// [START gae_cloudsql]

// Sample cloudsql demonstrates connection to a Cloud SQL instance from App Engine standard.
package main

import (
	"bytes"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"encoding/json"
	"strings"
	"crypto/subtle"
	"github.com/rs/cors"

	// MySQL library, comment out to use PostgreSQL.
	_ "github.com/go-sql-driver/mysql"
	// PostgreSQL Library, uncomment to use.
	// _ "github.com/lib/pq"
)

var db *sql.DB


const 
(
  ADMIN_USER = "admin"
  ADMIN_PASSWORD = "admin"
)

var ALLOW_SEARCH_FIELDS = []string { "name","type","district","address"}
	

var ALLOW_MATCH_FIELDS = []string {"id","verified"}

type Organization struct {
	Id int `json:"id"` 
	Name string `json:"name"` 
	OrgType string `json:"orgType"` 
	Website string `json:"website"` 
	Facebook string `json:"facebook"` 
	BrNumber string `json:"brNumber"` 
	Phone string `json:"phone"` 
	District string `json:"district"` 
	Address string `json:"address"` 
	ContactPersonName string `json:"contactPersonName"` 
	ContactPersonPhone string `json:"contactPersonPhone"` 
	ContactPersonRole string `json:"contactPersonRole"` 
	Email string `json:"email"` 
	TGId string `json:"tgId"` 
	PastExp string `json:"pastExp"` 
	Verified int `json:"verified"` 
	Lat float32 `json:"lat"` 
	Lng float32 `json:"lng"` 
	ShowContact int `json:"showContact"` 
	CustomRes string `json:"customResources"` 
	CustomServingTarget string `json:"customServingTarget"` 
	Resources []string `json:"resources"` 
	ServingTargets []string `json:"servingTargets"` 
}

type QtBuilding struct {
	Id int `json:"id"`
	ChiAddr string `json:"chiAddr"`
	EngAddr string `json:"engAddr"`
	District string `json:"district"`
	EndDate string `json:"endDate"`
	Lat float32 `json:"lat"` 
	Lng float32 `json:"lng"`
}

func enableCors(w *http.ResponseWriter) {
	(*w).Header().Add("Access-Control-Allow-Origin", "*")
}

func setupResponse(w *http.ResponseWriter, req *http.Request) {
	(*w).Header().Add("Access-Control-Allow-Origin", "*")
    (*w).Header().Add("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
    (*w).Header().Add("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
}

func main() {
	db = DB()
	mux := http.NewServeMux()
	mux.HandleFunc("/getOrg", BasicAuth(searchOrgHandler,"Please enter valid username and password"))
	mux.HandleFunc("/getQuaratineBuildingCount", qtBuildingCountHandler)
	mux.HandleFunc("/getQuaratineBuildingList", qtBuildingListHandler)
	mux.HandleFunc("/", indexHandler)
	
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Printf("Defaulting to port %s", port)
	}

	log.Printf("Listening on port %s", port)
	c := cors.New(cors.Options{
		AllowedOrigins: []string{"*", "https://flexible-frontend-dot-antivirus-center.appspot.com"},
		AllowCredentials: true,
		AllowedMethods: []string{"POST","GET","OPTIONS"},
		AllowedHeaders: []string{"Authorization"},
		// Enable Debugging for testing, consider disabling in production
		Debug: true,
	})
	handler := c.Handler(mux)
	if err := http.ListenAndServe(":"+port, handler); err != nil {
		log.Fatal(err)
	}
}

func BasicAuth(handler http.HandlerFunc, realm string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
	  user, pass, ok := r.BasicAuth()
	  if !ok || subtle.ConstantTimeCompare([]byte(user), []byte(ADMIN_USER)) != 1||subtle.ConstantTimeCompare([]byte(pass), []byte(ADMIN_PASSWORD)) != 1 {
		w.Header().Set("WWW-Authenticate", `Basic realm="`+realm+`"`)
		w.WriteHeader(401)
		w.Write([]byte("You are Unauthorized to access the application.\n"))
		return
	  }
	  handler(w, r)
	}
  }

// DB gets a connection to the database.
// This can panic for malformed database connection strings, invalid credentials, or non-existance database instance.
func DB() *sql.DB {
	var (
		connectionName = mustGetenv("CLOUDSQL_CONNECTION_NAME")
		user           = mustGetenv("CLOUDSQL_USER")
		dbName         = "antivirus" // NOTE: dbName may be empty
		password       = os.Getenv("CLOUDSQL_PASSWORD")      // NOTE: password may be empty
		socket         = os.Getenv("CLOUDSQL_SOCKET_PREFIX")
	)

	// /cloudsql is used on App Engine.
	if socket == "" {
		socket = "/cloudsql"
	}

	// MySQL Connection, comment out to use PostgreSQL.
	// connection string format: USER:PASSWORD@unix(/cloudsql/PROJECT_ID:REGION_ID:INSTANCE_ID)/[DB_NAME]
	dbURI := fmt.Sprintf("%s:%s@unix(%s/%s)/%s", user, password, socket, connectionName, dbName)
	conn, err := sql.Open("mysql", dbURI)

	// PostgreSQL Connection, uncomment to use.
	// connection string format: user=USER password=PASSWORD host=/cloudsql/PROJECT_ID:REGION_ID:INSTANCE_ID/[ dbname=DB_NAME]
	// dbURI := fmt.Sprintf("user=%s password=%s host=/cloudsql/%s dbname=%s", user, password, connectionName, dbName)
	// conn, err := sql.Open("postgres", dbURI)

	if err != nil {
		panic(fmt.Sprintf("DB: %v", err))
	}

	return conn
}

func Contains(a []string, x string) int {
	for i, n := range a {
			if x == n {
					return i
			}
	}
	return -1;
}

func returnOrgResult(rows *sql.Rows, err error, w http.ResponseWriter, r *http.Request){
	if err != nil {
		log.Printf("Could not query db: %v", err)
		http.Error(w, "Internal Error returnOrgResult", 500)
		return
	}
	defer rows.Close()
	buf := bytes.NewBufferString("")
	var organizations []Organization
	for rows.Next() {

		var customRes sql.NullString
		var customServingTarget sql.NullString
		var resourcesRaw sql.NullString
		var servingTargetRaw sql.NullString

		organization := Organization{}
		err = rows.Scan(&organization.Id,
						&organization.Name,
						&organization.OrgType,
						&organization.Website,
						&organization.Facebook,
						&organization.BrNumber,
						&organization.Phone,
						&organization.District,
						&organization.Address,
						&organization.ContactPersonName,
						&organization.ContactPersonPhone,
						&organization.ContactPersonRole,
						&organization.Email,
						&organization.TGId,
						&organization.PastExp,
						&organization.Verified,
						&organization.Lat,
						&organization.Lng,
						&organization.ShowContact,
						&customRes,
						&customServingTarget,
						&resourcesRaw,
						&servingTargetRaw);
		if err != nil {
			log.Printf("Scan Row error: %v", err)
			http.Error(w, "Internal Error", 500)
			return
		}
		organization.CustomRes = getValidString(customRes)
		organization.CustomServingTarget = getValidString(customServingTarget)
		organization.Resources = strings.Split(getValidString(resourcesRaw), ",")
		for i :=range organization.Resources {
			organization.Resources[i] = strings.TrimSpace(organization.Resources[i]);
		}
		organization.ServingTargets = strings.Split(getValidString(servingTargetRaw),",")		
		for j :=range organization.ServingTargets {
			organization.ServingTargets[j] = strings.TrimSpace(organization.ServingTargets[j]);
		}

		if organization.ShowContact < 1 {
			organization.Address = "";
			organization.Lat = 0;
			organization.Lng = 0;
		}


		organizations = append(organizations, organization)
		

	}
	orgsStr, _ := json.Marshal(organizations)
	fmt.Fprintf(buf, "%s", string(orgsStr))
	w.Write(buf.Bytes());
}


func qtBuildingCountHandler(w http.ResponseWriter, r *http.Request){
	rows, err := db.Query(`SELECT COUNT(*) from compulsory_quarantine`);

	if err != nil {
		log.Printf("Could not query db: %v", err)
		http.Error(w, "Internal Error qtBuildingCountHandler", 500)
		return
	}

	var count int = 0;
	buf := bytes.NewBufferString("")

	for rows.Next() {
		err := rows.Scan(&count);
		if err != nil {
			log.Printf("Scan Row error: %v", err)
			http.Error(w, "Internal Error", 500)
			return
		}
	}
	fmt.Fprintf(buf, "{ \"count\": %d}", count)
	w.Write(buf.Bytes());
}

func qtBuildingListHandler(w http.ResponseWriter, r *http.Request){
	start, _ := r.URL.Query()["start"]
	end, _ := r.URL.Query()["count"]
	district, _ := r.URL.Query()["district"]

	var startIndex string = "1"
	var limit string = "200"
	var districtStr string = ""
	if start != nil && len(start) > 0 {
		startIndex = start[0]
	}

	if end != nil && len(end) > 0 {
		limit = end[0]
	}

	if district != nil && len(district) > 0 {
		districtStr = district[0];
	}

	w.Header().Set("Content-Type", "text/json")

	var query string = `SELECT * from compulsory_quarantine WHERE id >= ? LIMIT `+limit

	var rows *sql.Rows;
	var err error;

	if len(districtStr) > 0 {
		query = `SELECT * from compulsory_quarantine WHERE id >= ? AND district = ? LIMIT `+limit
		rows, err = db.Query(query, startIndex, districtStr)
	}else{
		rows, err = db.Query(query, startIndex)
	}
	defer rows.Close()
	var buildings []QtBuilding

	var count int;

	buf := bytes.NewBufferString("")


	for rows.Next() {
		qtBuilding := QtBuilding{}
		err = rows.Scan(&qtBuilding.Id,
						&qtBuilding.ChiAddr,
						&qtBuilding.EngAddr,
						&qtBuilding.District,
						&qtBuilding.EndDate,
						&qtBuilding.Lat,
						&qtBuilding.Lng)
		if err != nil {
			log.Printf("Scan Row error: %v", err)
			http.Error(w, "Internal Error", 500)
			return
		}
		count++
		buildings = append(buildings, qtBuilding);
	}
	buildingsStr, _ := json.Marshal(buildings)
	fmt.Fprintf(buf, "{ \"data\": %s, \"count\": %d }", string(buildingsStr), count)
	w.Write(buf.Bytes());
}

func searchOrgHandler(w http.ResponseWriter, r *http.Request){
	searchKey, _ := r.URL.Query()["search_key"]
	searchValue, _ := r.URL.Query()["search_val"]
	if searchKey == nil || 
	len(searchKey) == 0 || 
	searchValue == nil || 
	len(searchValue) == 0 {
		getOrgHandler(w, r)
		return;
	}
	indexSearchField := Contains(ALLOW_SEARCH_FIELDS, searchKey[0])
	indexMatchField := Contains(ALLOW_MATCH_FIELDS, searchKey[0])  
	if indexSearchField < 0 {
		if indexMatchField < 0{
			getOrgHandler(w, r)
			return;
		}
	}
	w.Header().Set("Content-Type", "text/json")

	var compareStmt string
	var value string
	if indexSearchField >= 0 {
		compareStmt = "organization."+ALLOW_SEARCH_FIELDS[indexSearchField] + " LIKE ? "
		value = "%"+searchValue[0]+"%"
	}
	if indexMatchField >= 0 {
		compareStmt = "organization."+ALLOW_MATCH_FIELDS[indexMatchField] + " = ? "
		value = searchValue[0]
	}
	log.Printf("statement %s",compareStmt);
	rows, err := db.Query(`SELECT 
	organization.id,
	name,
	type,
	website,
	facebook,
	br_number, 
	phone,
	district,
	address,
	contact_person_name, 
	contact_person_phone,
	contact_person_role,
	email, 
	tg_id,
	past_exp,
	verified,
	lat,
	lng,
	show_contact,
	other_resource_type.type_content as custom_res, 
	other_serving_target.target_name as custom_target, 
	(SELECT GROUP_CONCAT(DISTINCT org_resource_type.res_id ORDER BY org_resource_type.res_id SEPARATOR ', ') FROM org_resource_type WHERE org_resource_type.org_id = organization.id) as resources, 
	(SELECT GROUP_CONCAT(DISTINCT org_serving_target.serving_target_id ORDER BY org_serving_target.serving_target_id SEPARATOR ', ') FROM org_serving_target WHERE org_serving_target.org_id = organization.id) as serving_targets 
	FROM organization 
	LEFT JOIN other_resource_type on organization.id = other_resource_type.org_id 
	LEFT JOIN other_serving_target on organization.id = other_serving_target.org_id 
	WHERE `+ compareStmt + ` AND verified = 1 ORDER BY organization.id`, value);

	returnOrgResult(rows, err, w, r);
}

func getOrgHandler(w http.ResponseWriter, r *http.Request){
	w.Header().Set("Content-Type", "text/json")
	rows, err := db.Query(`SELECT 
	organization.id,
	name,
	type,
	website,
	facebook,
	br_number, 
	phone,
	district,
	address,
	contact_person_name, 
	contact_person_phone,
	contact_person_role,
	email, 
	tg_id,
	past_exp,
	verified,
	lat,
	lng,
	show_contact,
	other_resource_type.type_content as custom_res, 
	other_serving_target.target_name as custom_target, 
	(SELECT GROUP_CONCAT(DISTINCT org_resource_type.res_id ORDER BY org_resource_type.res_id SEPARATOR ', ') FROM org_resource_type WHERE org_resource_type.org_id = organization.id) as resources, 
	(SELECT GROUP_CONCAT(DISTINCT org_serving_target.serving_target_id ORDER BY org_serving_target.serving_target_id SEPARATOR ', ') FROM org_serving_target WHERE org_serving_target.org_id = organization.id) as serving_targets 
	FROM organization 
	LEFT JOIN other_resource_type on organization.id = other_resource_type.org_id 
	LEFT JOIN other_serving_target on organization.id = other_serving_target.org_id 
	WHERE verified = 1
	ORDER BY organization.id`);

	returnOrgResult(rows, err, w, r);
	
}

func getValidString(str sql.NullString) string{
	if str.Valid {
		return str.String
	}else{
		return ""
	}
}

// indexHandler responds to requests with our list of available databases.
func indexHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "text/json")

	rows, err := db.Query("SHOW DATABASES")
	if err != nil {
		log.Printf("Could not query db: %v", err)
		http.Error(w, "Internal Error", 500)
		return
	}
	defer rows.Close()

	buf := bytes.NewBufferString("Databases:\n")
	/*for rows.Next() {
		var dbName string
		if err := rows.Scan(&dbName); err != nil {
			log.Printf("Could not scan result: %v", err)
			http.Error(w, "Internal Error", 500)
			return
		}
		fmt.Fprintf(buf, "- %s\n", dbName)
	}*/
	fmt.Fprintf(buf, "API Running")
	w.Write(buf.Bytes());
}

func mustGetenv(k string) string {
	v := os.Getenv(k)
	if v == "" {
		log.Panicf("%s environment variable not set.", k)
	}
	return v
}

// [END gae_cloudsql]
