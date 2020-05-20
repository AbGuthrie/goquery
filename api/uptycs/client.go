package uptycs

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/AbGuthrie/goquery/v2/hosts"
	"github.com/AbGuthrie/goquery/v2/models"
	"github.com/dgrijalva/jwt-go"
	id "github.com/google/uuid"
	"github.com/tidwall/gjson"
	"golang.org/x/net/context"
)

type uptycsConfig struct {
	Domain     string    `json:"domain"`
	UserID     string    `json:"userId"`
	Active     bool      `json:"active"`
	UpdatedAt  time.Time `json:"updatedAt"`
	Secret     string    `json:"secret"`
	CustomerID string    `json:"customerId"`
	ID         string    `json:"ID"`
	CreatedAt  time.Time `json:"createdAt"`
	Key        string    `json:"key"`
}

type UptycsAPI struct {
	Authed        bool
	defaultReqObj *http.Request
	httpClient    *http.Client
	uptCfg        *uptycsConfig
}

var (
	scheduleQueryTemplate = `{"type":"realtime","query":"%s","filtering":{"filters":{"id":{"equals":"%s"}}}}`
	db                    map[string]string
)

/*
	helper functions
*/

func credentials() (*uptycsConfig, error) {
	fmt.Println("Enter path to Uptycs credentials file")
	confFilePath, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("Error reading filepath value: %s", err)
	}
	confFilePath = strings.TrimSpace(confFilePath)
	confFileContent, err := ioutil.ReadFile(confFilePath)
	if err != nil {
		return nil, fmt.Errorf("Error reading Uptycs credential file content: %s", err)
	}
	cfg := &uptycsConfig{}
	if err := json.Unmarshal(confFileContent, &cfg); err != nil {
		return nil, fmt.Errorf("Error parsing Uptycs credential file: %s ", err)
	}
	return cfg, nil
}

// CreateUptycsAPI creates an authenticated instance of the `UptycsAPI` object
func CreateUptycsAPI() (models.GoQueryAPI, error) {
	retVal := UptycsAPI{}
	cfg, err := credentials()
	if err != nil {
		return retVal, err
	}
	retVal.httpClient = &http.Client{}
	retVal.defaultReqObj, err = http.NewRequest(
		"GET",
		fmt.Sprintf(
			"https://%s.uptycs.io/public/api/customers/%s",
			cfg.Domain,
			cfg.CustomerID,
		),
		nil,
	)
	if err != nil {
		return retVal, fmt.Errorf(
			"Error instantiating HTTP request object: %s",
			err,
		)
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"iss": cfg.Key,
	}).SignedString([]byte(cfg.Secret))
	if err != nil {
		return retVal, fmt.Errorf("Error creating bearer token: %s", err)
	}
	retVal.defaultReqObj.Header.Add(
		"Authorization",
		fmt.Sprintf("Bearer %s", token),
	)
	retVal.Authed = true
	db = make(map[string]string)
	return retVal, nil
}

/*
	interface implementation
*/

func (u UptycsAPI) doHTTPReq(req *http.Request) (string, error) {
	// FOR DEBUG
	/*
		reqBytes, err := httputil.DumpRequest(req, true)
		if err != nil {
			return "", nil
		}
		fmt.Println("--- HTTP REQ ---")
		fmt.Println(string(reqBytes))
	*/

	resp, err := u.httpClient.Do(req)
	if err != nil {
		return "", nil
	}

	// FOR DEBUG
	/*
		respBytes, err := httputil.DumpResponse(resp, true)
		if err != nil {
			return "", nil
		}
		fmt.Println("--- HTTP RESP ---")
		fmt.Println(string(respBytes))
		fmt.Println("")
	*/

	respData, err := ioutil.ReadAll(resp.Body)
	return string(respData), err
}

func (u UptycsAPI) getAssetInfo(uuid string) (string, error) {
	req := u.defaultReqObj.Clone(context.TODO())
	req.URL.Path = fmt.Sprintf(
		"%s/assets/%s", req.URL.Path, uuid,
	)
	return u.doHTTPReq(req)
}

func (u UptycsAPI) getUsers(uuid string) (string, error) {
	req := u.defaultReqObj.Clone(context.TODO())
	req.URL.Path = fmt.Sprintf(
		"%s/assets/%s/user", req.URL.Path, uuid,
	)
	return u.doHTTPReq(req)
}

func (u UptycsAPI) CheckHost(uuid string) (hosts.Host, error) {
	retVal := hosts.Host{
		UUID:             uuid,
		CurrentDirectory: "/",
	}
	if !u.Authed {
		return retVal, errors.New("Error, UptycsAPI object is not yet initialized")
	}
	queryResult, err := u.getAssetInfo(uuid)
	if err != nil {
		return retVal, err
	}
	errMsg := gjson.Get(queryResult, "error").String()
	if len(errMsg) > 0 {
		return retVal, fmt.Errorf("Uptycs API error: %s", errMsg)
	}
	retVal.ComputerName = gjson.Get(queryResult, "hostName").String()
	retVal.Platform = fmt.Sprintf("%s %s - %s %s",
		gjson.Get(queryResult, "hardwareVendor").String(),
		gjson.Get(queryResult, "hardwareModel").String(),
		gjson.Get(queryResult, "os").String(),
		gjson.Get(queryResult, "os_key").String(),
	)
	retVal.Version = gjson.Get(queryResult, "osqueryVersion").String()
	queryResult, err = u.getUsers(uuid)
	if err != nil {
		return retVal, err
	}
	/*
		this is not a good way to get the current user, but it seemed like a
		better idea at the time than making the following realtime query:
			SELECT user || ' (' || tty || ')' FROM logged_in_users
			WHERE type = 'user'
			and upt_asset_id = '%s'
			ORDER BY time LIMIT 1`
	*/
	retVal.Username = gjson.Get(queryResult, `items.#(uid=="501").username`).String()
	return retVal, nil
}

/*
	This is a dirty hack that will make memory expand per query due to its design.
	Uptycs currently does not have a way to schedule queries on an individual host.
	So, instead, we make a realtime query, and store the result in a map with a UUID
	as a key, then return that as the "result".
*/
func (u UptycsAPI) ScheduleQuery(uuid string, query string) (string, error) {
	if !u.Authed {
		return "", errors.New("Error, UptycsAPI object is not yet initialized")
	}
	query = strings.ReplaceAll(query, `"`, `\"`)
	finalBodyString := fmt.Sprintf(scheduleQueryTemplate, query, uuid)
	req := u.defaultReqObj.Clone(context.TODO())
	req.URL.Path = fmt.Sprintf("%s/assets/query", req.URL.Path)
	req.Method = "POST"
	req.Body = ioutil.NopCloser(bytes.NewBufferString(finalBodyString))
	req.Header.Add("Content-Length", fmt.Sprintf("%d", len(finalBodyString)))
	req.Header.Add("Content-Type", "application/json")

	result, err := u.doHTTPReq(req)
	if err != nil {
		return "", fmt.Errorf("Error making scheduled query: %s", err)
	}
	queryUUID := id.New().String()
	db[queryUUID] = result
	hosts.AddQueryToHost(uuid, hosts.Query{
		Name: queryUUID,
		SQL:  query,
	})
	return queryUUID, nil
}

func (u UptycsAPI) FetchResults(queryName string) ([]map[string]string, string, error) {
	retVal := []map[string]string{}
	queryResult, ok := db[queryName]
	if !ok {
		return retVal, "", errors.New("Query not found")
	}
	for _, item := range gjson.Get(queryResult, "items").Array() {
		row := make(map[string]string)
		for key, value := range item.Map() {
			row[key] = value.String()
		}
		retVal = append(retVal, row)
	}
	return retVal, "", nil
}
