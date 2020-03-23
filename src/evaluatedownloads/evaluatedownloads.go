package evaluatedownloads

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"encoding/base64"
	"log"
	"net/url"
	"encoding/json"
	"strings"

	"jfrog-test/src/common"
)


func basicAuth(username, password string) string {
	auth := fmt.Sprintf("%s:%s", username, password)
	//log.Println(auth)
	return base64.StdEncoding.EncodeToString([]byte(auth))
}

func getSecurityToken() string {
	// getting username and password from environemnt variable is temporary workaround
	userName := os.Getenv("JFROG_ARTIFACTORY_USER");
	password := os.Getenv("JFROG_ARTIFACTORY_PASSWORD");
	if (userName == "" || password == "") {
	  panic("Failed to get credentials");
	}

	// curl -u <user>:<pwd> -iX POST http://34.71.214.77/artifactory/api/security/token
	//  -d'username=<user>' -d'scope=member-of-groups:readers'

	data := url.Values{}
	data.Set("username", userName)
	data.Add("scope", "member-of-groups:readers")

	url := fmt.Sprintf("%s%s", os.Getenv("JFROG_ARTIFACTORY_URL"), "/api/security/token");

	req, err := http.NewRequest("POST", url, bytes.NewBufferString(data.Encode()))
	req.Header.Add("Authorization","Basic " + basicAuth(userName, password))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	var tokenInfo map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&tokenInfo)
	accessToken := tokenInfo["access_token"].(string)
	//log.Println(accessToken)
	return accessToken;
}

func getArtifactsList(token string, repoKey string) []interface{} {
	if repoKey == "" {
		panic("Invalid Params")
	}

	// curl -iX POST http://34.71.214.77/artifactory/api/search/aql
	// -H "Content-Type: text/plain"  -H "Authorization: Bearer <token>" -d'items.find({"repo":{"$eq":"jcenter-cache"}})'

	url := fmt.Sprintf("%s%s", os.Getenv("JFROG_ARTIFACTORY_URL"), "/api/search/aql")
	body := fmt.Sprintf("items.find({\"repo\":{\"$eq\":\"%s\"}})", repoKey)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(body)))
	req.Header.Set("Content-Type", "text/plain")
	if token == "" {
		userName := os.Getenv("JFROG_ARTIFACTORY_USER")
		password := os.Getenv("JFROG_ARTIFACTORY_PASSWORD")
		req.Header.Add("Authorization","Basic " + basicAuth(userName, password))
	} else {
		bearerToken := fmt.Sprintf("%s%s", "Bearer ", token)
		req.Header.Set("Authorization", bearerToken)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	var artifacts map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&artifacts)
	// for pretty print
	//artifactsList,_ := json.MarshalIndent(artifacts["results"], "", "  ")
	//log.Println(string(artifactsList))
	return artifacts["results"].([]interface{})
}


func getFileStats(token string, fileURI string) int64 {
	if fileURI == "" {
		panic("Invalid Params")
	}

	// curl -i GET http://34.71.214.77/artifactory/api/storage/<repoKey>/<repoPath>/<fileName>?stats
	//  -H "Authorization: Bearer <token>"

	url := fmt.Sprintf("%s%s%s?stats", os.Getenv("JFROG_ARTIFACTORY_URL"), "/api/storage/", fileURI)

	req, err := http.NewRequest("GET", url, nil)
	if token == "" {
		userName := os.Getenv("JFROG_ARTIFACTORY_USER")
		password := os.Getenv("JFROG_ARTIFACTORY_PASSWORD")
		req.Header.Add("Authorization","Basic " + basicAuth(userName, password))
	} else {
		bearerToken := fmt.Sprintf("%s%s", "Bearer ", token)
		req.Header.Set("Authorization", bearerToken)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	var fileDownloadInfo map[string]interface{}
	var downloadCount int64 = -1
	json.NewDecoder(resp.Body).Decode(&fileDownloadInfo)
	if fileDownloadInfo != nil {
		// for pretty print
		//fileDownload,_ := json.MarshalIndent(fileDownloadInfo, "", "  ")
		//log.Println(string(fileDownload))
		downloadCount = fileDownloadInfo["downloadCount"].(int64)
	}

	return downloadCount
}

func processPopularDownloads(topK int, accessToken string, fileURI string, maxHeap *common.MaxHeap)  {
	if accessToken == "" {
		panic("Invalid Params")
	}

	log.Printf("File stats for %s ...\n", fileURI)
	downloadCount := getFileStats(accessToken, fileURI)
	fileDownloadCountObj := common.NewFileDownloadCount(fileURI, downloadCount)
	maxHeap.Insert(fileDownloadCountObj)
}

// EvaluatePopularDownloads ...
// TODO: have to support multiple repoTypes, fileTypes
//func EvaluatePopularDownloads(topK uint64, repoKey string, typeRepoFile map[string][])  {
func EvaluatePopularDownloads(topK int, repoKey string, typeRepoNFileTypes map[string][]string) {
	accessToken := getSecurityToken()
	artifactsList := getArtifactsList(accessToken, repoKey)
	//TODO: create maxHeap for each fileType in fileTypes
	filetypeMaxHeap := common.NewMaxHeap(topK)
	for _, artifact := range artifactsList {
		artifactInfo := artifact.(map[string]interface{})
		artifactPath := artifactInfo["path"].(string)
		//TODO: optimization: get repoType from artifactPath and look in typeRepoNFileTypes
		for repoType, fileTypes := range typeRepoNFileTypes {
			if strings.Contains(artifactPath, repoType) {
				artficatName := artifactInfo["name"].(string)
				//TODO: optimization fileTypes has to be in set
				for _, fileType := range fileTypes {
					if strings.HasSuffix(artficatName, fileType) {
						fileURI := fmt.Sprintf("%s/%s/%s", repoKey, artifactPath, artficatName)
						//TODO: get correspoinding maxHeap reciever (filetypeMaxHeap)
						go processPopularDownloads(topK, accessToken, fileURI, filetypeMaxHeap)
					}
				}
			}
		}
	}
	// TODO: save maxHeap into Redis
	// UpdatePopularDownloadsIntoRedis
}


func evaluatedownloadsDemo() {
	repoKey := "jcenter-cache"
	topK := 2
	var typeRepoNFileTypes map[string][]string
	typeRepoNFileTypes["maven"] = []string{"jar", "pom"}

	EvaluatePopularDownloads(topK, repoKey, typeRepoNFileTypes)
}
