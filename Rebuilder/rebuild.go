package Rebuilder

import (
	"DuckBox/Define"
	"DuckBox/conf"
	"DuckBox/controllers"
	"fmt"
	"github.com/meilisearch/meilisearch-go"
	"strings"
)

func RebuildVideoIndex() error {

	data, ids, err := GetVideosFromMySQL()
	if err != nil {
		//log
		controllers.Log("获取数据库中文档失败,错误信息:", err.Error(), controllers.LogError)
		return err
	}

	existingDocumentsId, err := GetExistingDocumentsID(controllers.MeiliSearchIndexVideo)
	if err != nil {
		//log
		controllers.Log("获取现有索引失败", err.Error(), controllers.LogError)
		return err
	}

	//获取现有文档与数据库文档的差集
	d3 := difference(ids, existingDocumentsId)
	if len(d3) <= 0 {
		//log
		controllers.Log("更新文档数量", "没有文档需要更新", controllers.LogInfo)
		return nil
	}
	//log
	controllers.Log("更新文档数量", fmt.Sprintf("有%d个文档需要更新", len(d3)), controllers.LogInfo)
	//insert document
	var insertDocument []Define.VideoSearchModel
	//遍历差集 数组，生成插入文档
	for _, s := range d3 {
		if val, exits := data[s]; exits == true {
			insertDocument = append(insertDocument, Define.VideoSearchModel{
				Title:       val.Title,
				Description: val.Description,
				Id:          val.Id,
				Thumb:       val.Thumb,
				CreateTime:  val.Published,
				Viewer:      val.Viewer,
				CreatorId:   val.Applicant.Id,
				CreatorName: val.Applicant.Username,
			})
		}
	}
	//log
	controllers.Log("更新文档具体ID，正在与MeiliSearch连接", strings.Join(d3, ","), controllers.LogInfo)
	//与meiliSearch 连接
	client := meilisearch.NewClient(meilisearch.ClientConfig{
		Host:   conf.AppConfig.MeiliSearch.ToHost(),
		APIKey: conf.AppConfig.MeiliSearch.ApiKey,
	})
	//插入文档
	r1, err := client.Index(controllers.MeiliSearchIndexVideo).AddDocuments(&insertDocument, "id")
	if err != nil {
		//log
		controllers.Log("更新搜索索引失败", err.Error(), controllers.LogError)
		return err
	}
	//log
	controllers.Log("成功更新,成功状态信息:", string(r1.Status), controllers.LogInfo)
	return nil
}

func RebuildProgramIndex() error {

	data, ids, err := GetProgramsFromMySQL()
	if err != nil {
		controllers.Log("获取数据库中文档失败,错误信息:", err.Error(), controllers.LogError)
		return err
	}

	existingDocuments, err := GetExistingDocumentsID(controllers.MeiliSearchIndexProgram)
	if err != nil {
		controllers.Log("获取现有索引失败", err.Error(), controllers.LogError)
		return err
	}

	d3 := difference(ids, existingDocuments)
	if len(d3) <= 0 {
		//log
		controllers.Log("更新文档数量", "没有文档需要更新", controllers.LogInfo)
		return nil
	}

	//log
	controllers.Log("更新文档数量", fmt.Sprintf("有%d个文档需要更新", len(d3)), controllers.LogInfo)

	var insertDocument []Define.ProgramSearchModel

	//遍历差集 数组，生成插入文档
	for _, s := range d3 {
		if val, exits := data[s]; exits == true {
			insertDocument = append(insertDocument, Define.ProgramSearchModel{
				Title:        val.Title,
				ShowSubtitle: val.ShowSubTitle,
				Description:  val.Description,
				Id:           val.Id,
				Poster:       val.Poster,
				CreateTime:   val.CreateTime,
				Viewer:       val.Viewer,
				Subscriber:   val.Subscriber,
				CreatorId:    val.Applicant.Id,
				CreatorName:  val.Applicant.Username,
			})
		}
	}
	if len(insertDocument) <= 0 {
		controllers.Log("没有文档更新", "", controllers.LogInfo)
		return nil
	}

	controllers.Log("更新文档具体ID，正在与MeiliSearch连接", strings.Join(d3, ","), controllers.LogInfo)
	//与meiliSearch 连接
	client := meilisearch.NewClient(meilisearch.ClientConfig{
		Host:   conf.AppConfig.MeiliSearch.ToHost(),
		APIKey: conf.AppConfig.MeiliSearch.ApiKey,
	})
	//插入文档
	r1, err := client.Index(controllers.MeiliSearchIndexProgram).AddDocuments(&insertDocument, "id")
	if err != nil {
		//log
		controllers.Log("更新搜索索引失败", err.Error(), controllers.LogError)
		return err
	}

	//log
	controllers.Log("成功更新,成功状态信息:", string(r1.Status), controllers.LogInfo)
	return nil
}
