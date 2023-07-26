package Rebuilder

import (
	"DuckBox/controllers"
	"DuckBox/models"
	"fmt"
	"github.com/astaxie/beego/orm"
	"github.com/meilisearch/meilisearch-go"
	"math"
	"strings"
)

func GetVideosFromMySQL() (map[string]*models.Video, []string, error) {

	var (
		o           = orm.NewOrm()
		result      []*models.Video
		responseIds []string
		response    = make(map[string]*models.Video)
	)

	if _, err := o.QueryTable(&models.Video{}).Filter("State", 1).All(&result); err != nil {
		return nil, nil, err
	}

	for _, video := range result {
		//load creator information
		if _, err := o.LoadRelated(video, "Applicant"); err != nil {
			return nil, nil, err
		}
		//
		responseIds = append(responseIds, video.Id)
		//if in response
		if _, exist := response[video.Id]; exist == false {
			response[video.Id] = video
		}

	}

	return response, responseIds, nil
}

func GetProgramsFromMySQL() (map[string]*models.Program, []string, error) {

	var (
		o           = orm.NewOrm()
		result      []*models.Program
		responseIds []string
		response    = make(map[string]*models.Program)
	)

	if _, err := o.QueryTable(&models.Program{}).Filter("State", 1).All(&result); err != nil {
		return nil, nil, err
	}

	for _, program := range result {
		//
		responseIds = append(responseIds, program.Id)
		//response
		if _, exist := response[program.Id]; exist == false {
			response[program.Id] = program
		}
	}

	return response, responseIds, nil
}

func GetExistingDocumentsID(indexName string) ([]string, error) {

	client := meilisearch.NewClient(
		meilisearch.ClientConfig{
			Host:   controllers.MeiliSearchHost,
			APIKey: controllers.MeiliSearchAPIKey,
		},
	)

	indexes, err := client.GetIndexes(&meilisearch.IndexesQuery{
		Limit:  20,
		Offset: 0,
	})
	if err != nil {
		return nil, err
	}

	var isExistIndex = false

	for _, result := range indexes.Results {

		if strings.Compare(result.UID, indexName) == 0 {
			isExistIndex = true
		}
	}

	if isExistIndex == false {

		_, err := client.CreateIndex(&meilisearch.IndexConfig{
			Uid:        indexName,
			PrimaryKey: "id",
		})
		if err != nil {
			return nil, err
		}
	}

	var result meilisearch.DocumentsResult
	var query meilisearch.DocumentsQuery
	var response []string

	query.Limit = 200
	query.Offset = 0
	query.Fields = []string{"id"}

	err = client.Index(indexName).GetDocuments(&query, &result)
	if err != nil {
		//log
		return nil, err
	}

	//
	for _, m := range result.Results {
		response = append(response, fmt.Sprintf("%v", m["id"]))
	}

	if result.Total > 200 {
		//获取页数
		var page = int(math.Round(float64(result.Total / 200)))
		//循环获取数据
		for i := 0; i < page; i++ {
			//获取每一页数据
			err = client.Index(indexName).GetDocuments(&meilisearch.DocumentsQuery{
				Offset: int64((page - 1) * 200),
				Limit:  200,
			}, &result)
			//假如出现问题
			if err != nil {
				return nil, err
			}
			//添加到
			for _, m := range result.Results {
				response = append(response, fmt.Sprintf("%v", m["id"]))
			}
		}
	}
	return response, nil
}

func difference(d1 []string, d2 []string) []string {

	//var a = make(map[string]bool)
	var response []string

	for _, s := range d1 {
		var re = false
		//
		for _, s2 := range d2 {
			if strings.Compare(s, s2) == 0 {
				re = true
				break
			}
		}
		if re == false {
			response = append(response, s)
		}
	}

	return response
}
