package common

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/go-redis/redis/v7"
)

// InitRedisSession ...
func InitRedisSession() *redis.Client {
	redisServer := os.Getenv("REDIS_URL")
	if redisServer == "" {
		return nil
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr:     redisServer,
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	pong, err := redisClient.Ping().Result()
	fmt.Println(pong, err)
	if err != nil {
		//Output: dial tcp 127.0.0.1:6379: connect: connection refused
		log.Println("Failed to connect Redis")
	}
	// Output: PONG <nil>
	return redisClient
}

// ReadResourceCountFromRedis ...
func ReadPopularDownloadsFromRedis(topK int, repoKey string, ftypeRepoNFileTypes map[string][]string,
	redisClient *redis.Client) map[string]map[string][]interface{} {
	// Reading PopularDownloads from Redis
	topKDownloadsList := map[string]map[string][]interface{}{}
	repokeyrepoTypefileTypeMap := map[string]map[string]map[string][]interface{}{}

	if redisClient == nil {
		return topKDownloadsList
	}

	if val, err := redisClient.Get("PopularDownloads").Result(); err == nil {
		// "PopularDownloads" key exists
		if err = json.Unmarshal([]byte(val), &repokeyrepoTypefileTypeMap); err != nil {
			log.Println("Failed to parse PopularDownload json")
		} else {
			// {repoKey: {repoType: fileType:[fileDownloadCount1, fileDownloadCount2]}}
			topKDownloadsList = repokeyrepoTypefileTypeMap[repoKey]
		}
	}
	return topKDownloadsList
}

// UpdatePopularDownloadsIntoRedis ...
func UpdatePopularDownloadsIntoRedis(repoKey string,
	repoTypeFiletypePopularDownloadsMap map[string]map[string][]interface{},
	redisClient *redis.Client) bool {
	if redisClient == nil {
		return false
	}

	update := true
	repoKeyRepotypeFiletypePopularDownloadsMap := map[string]map[string]map[string][]interface{}{}

	if val, err := redisClient.Get("PopularDownloads").Result(); err != nil {
		if err != redis.Nil {
			// redis not working ?
			update = false
			log.Println("Failed to get PopularDownloads from Redis")
		}
	} else {
		// "PopularDownloads" key exists
		if err = json.Unmarshal([]byte(val), &repoTypeFiletypePopularDownloadsMap); err != nil {
			log.Println("Failed to parse PopularDownloads json")
		}
	}

	if update {
		repoKeyRepotypeFiletypePopularDownloadsMap[repoKey] = repoTypeFiletypePopularDownloadsMap

		if repoTypeFiletypePopularDownloadsJSON, err := json.Marshal(repoTypeFiletypePopularDownloadsMap); err == nil {
			if err := redisClient.Set("PopularDownloads", string(repoTypeFiletypePopularDownloadsJSON), 0).Err(); err != nil {
				log.Println("Failed to update repoTypeFiletypePopularDownloadsMap into Redis")
				update = false
			}
		} else {
			update = false
		}
	}

	return update
}
