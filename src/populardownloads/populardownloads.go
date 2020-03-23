// Package populardownloads ...
package populardownloads

import (
	"log"
	"errors"
	"net/http"
	"encoding/json"

	"github.com/go-redis/redis/v7"

	"jfrog-test/src/common"
	"jfrog-test/src/evaluatedownloads"
)

// PopularDownloads ...
type PopularDownloads struct{}

// PDBody ...
type PDBody struct {
	topK int
	repoKey string
	typeRepoNFileTypes map[string][]string
}

var redisClient *redis.Client
// Init ...
func (pd *PopularDownloads) Init() {
	redisClient = common.InitRedisSession()
}

// ListTopDownloads ...
func (pd *PopularDownloads) ListTopDownloads(w http.ResponseWriter, r *http.Request) {
	var pdbody PDBody
	err := common.DecodeJSONBody(w, r, &pdbody)
    if err != nil {
        var mr *common.MalformedRequest
        if errors.As(err, &mr) {
            http.Error(w, mr.Msg, mr.Status)
        } else {
            log.Println(err.Error())
            http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
        }
        return
    }
    log.Printf("PDBody: %+v", pdbody)

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")


	topKDownloadsList := map[string]map[string][]interface{}{}

	// output format: {repoKey: {repoType: {fileType: []}}}
	// {"jcenter-cache": {"maven": {"jar": [{"name": "xy*.jar", downloadCount: 3}, {"name": "xy*.jar", downloadCount: 1}]}}}
	topKDownloadsList = common.ReadPopularDownloadsFromRedis(pdbody.topK, pdbody.repoKey, pdbody.typeRepoNFileTypes, redisClient)
	if len(topKDownloadsList) == 0 {
		go evaluatedownloads.EvaluatePopularDownloads(pdbody.topK, pdbody.repoKey, pdbody.typeRepoNFileTypes)
	}
	topKDownloadsListMap := map[string]map[string]map[string][]interface{}{}
	topKDownloadsListMap[pdbody.repoKey] = topKDownloadsList
	topKDownloadsListjson, err := json.Marshal(topKDownloadsListMap)

	w.Write([]byte(topKDownloadsListjson))
}
