package Rebuilder

import (
	"DuckBox/DataModel"
	"DuckBox/Define"
	"DuckBox/conf"
	"DuckBox/controllers"
	"context"
	"errors"
	"fmt"
	"github.com/meilisearch/meilisearch-go"
	"github.com/zhenghaoz/gorse/client"
	"sync"
	"time"
)

type Builder struct {
	SubmitTask func(func(*sync.WaitGroup))
}

func (b *Builder) BuildVideoIndex() error {

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
	var num, limit = len(insertDocument), 100
	if num <= 0 {
		return nil
	}
	var indexName = controllers.MeiliSearchIndexVideo
	// Calculate page number
	var p, p1 = num / limit, num % limit
	if p1 != 0 {
		p = p + 1
	}

	//
	if b.SubmitTask == nil {
		//if the total bigger than limit
		if num > limit {
			return errors.New("the quantity exceeds the limit, please set the SubmitTask Function")
		}
		//return error message
		return send2MeiliSearch[Define.VideoSearchModel](indexName, insertDocument)
	} else {
		for i := 0; i < p; i++ {

			var start, end = i * limit, (i + 1) * limit
			if end > num {
				end = num
			}
			if start > num {
				break
			}
			var insertData = insertDocument[start:end]
			//submit task
			b.SubmitTask(func(group *sync.WaitGroup) {

				err = send2MeiliSearch[Define.VideoSearchModel](indexName, insertData)
				if err != nil {
					//record the error message
					controllers.Log("send the doc to meili search", err.Error(), controllers.LogError)
					return
				}
				defer group.Done()
			})
		}
	}
	return nil
}

func (b *Builder) BuildProgramIndex() error {

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

	var num, limit = len(insertDocument), 100
	//if the number is zero
	if num <= 0 {
		return nil
	}
	// Calculate page number
	var p, p1 = num / limit, num % limit
	if p1 != 0 {
		p = p + 1
	}
	//index name is program
	var indexName = controllers.MeiliSearchIndexProgram
	//if the not set the submit task callback function
	if b.SubmitTask == nil {
		if num > limit {
			return errors.New("the quantity exceeds the limit, please set the SubmitTask Function")
		}
		//return the error message
		return send2MeiliSearch[Define.ProgramSearchModel](indexName, insertDocument)
	} else {
		//loop to submit task send document to meili search
		for i := 0; i < p; i++ {
			//callback function
			var start, end = i * limit, (i + 1) * limit
			if end > num {
				end = num
			}
			if start > num {
				break
			}
			var insertData = insertDocument[start:end]
			//submit task
			b.SubmitTask(func(g *sync.WaitGroup) {
				//use the core code
				err := send2MeiliSearch[Define.ProgramSearchModel](indexName, insertData)
				if err != nil {
					//record the error message
					controllers.Log("send the doc to meilisearch", err.Error(), controllers.LogError)
					return
				}
				defer g.Done()
			})
		}
	}
	return nil
}

func (b *Builder) BuildRecommendItems() error {

	video, ids, err := GetVideosFromMySQL()
	if err != nil {
		return err
	}

	program, pIds, err := GetProgramsFromMySQL()
	if err != nil {
		return err
	}
	//
	id, err := GetExistItemsId()
	if err != nil {
		controllers.Log("获取推荐文档失败!", err.Error(), controllers.LogError)
		return err
	}

	episodeMap, err := GetEpisodeMap()
	if err != nil {
		return err
	}

	var vi, pi = id["video"], id["program"]
	var insertItem []client.Item

	re := difference(ids, vi)

	for _, s := range re {

		if val, exist := video[s]; !exist {
			continue
		} else {
			// By default, classified as video
			Category := []string{Define.GorseCategoryVideo}
			//check if in the episode list
			if _, exist = episodeMap[val.Id]; exist {
				//if is episode
				Category = []string{Define.GorseCategoryEpisode}
			}
			//loop to get a tag and make label list
			var labels []string
			for _, tag := range val.Tags {
				//append to the label list
				labels = append(labels, tag.Name)
			}
			insertItem = append(insertItem, client.Item{
				ItemId:     Define.MakeItemId("video", val.Id),
				IsHidden:   val.State != DataModel.VideoStatusNormal,
				Labels:     labels,
				Categories: Category,
				Timestamp:  time.Unix(val.Published, 0).Format("2006-01-02"),
				Comment:    val.Title,
			})
		}
	}

	//make a list of insert program item
	pre := difference(pIds, pi)

	for _, s := range pre {
		if val, exist := program[s]; !exist {
			continue
		} else {
			var Labels []string
			//loop to get tag and make labels
			for _, tag := range val.Tags {
				Labels = append(Labels, tag.Name)
			}
			insertItem = append(insertItem, client.Item{
				ItemId:     Define.MakeItemId("program", val.Id),
				IsHidden:   val.State != DataModel.VideoStatusNormal,
				Labels:     Labels,
				Categories: []string{Define.GorseCategoryProgram},
				Timestamp:  time.Unix(val.CreateTime, 0).Format("2006-01-02"),
				Comment:    val.Title,
			})
		}
	}

	//Maximum number of submissions per time
	var limit, num = 1000, len(insertItem)

	if num <= 0 {
		return nil
	}
	//if the submit task callback func is not empty
	if b.SubmitTask != nil {

		if len(insertItem) < limit {
			b.SubmitTask(func(group *sync.WaitGroup) {
				err = AddItemsToGorse(insertItem)
				if err != nil {
					return
				}
				group.Done()
			})
		} else {
			var p, p1 = num / limit, num % limit
			if p1 != 0 {
				p = p + 1
			}
			for i := 0; i < p; i++ {
				var start, end = i * limit, (i + 1) * limit
				//if the end point bigger than total
				//30666 i=306 end=305
				if end > num {
					end = num
				}
				//if the start point bigger than total
				if start > num {
					break
				}
				var insertData = insertItem[start:end]

				b.SubmitTask(func(g *sync.WaitGroup) {
					if err := AddItemsToGorse(insertData); err != nil {
						//record the error message
						controllers.Log("send to gorse", err.Error(), controllers.LogError)
						return
					}
					//done
					defer g.Done()
				})
			}
		}
	} else {
		if num > limit {
			return errors.New("the quantity exceeds the limit, please set the SubmitTask Function")
		}
		//return error
		return AddItemsToGorse(insertItem)
	}

	return nil
}

func (g *Builder) BuildRecommendCustomerItem() error {

	customerId, customerMap, err := GetExistCustomerIdFromDB()
	if err != nil {
		return err
	}

	existCustomer, existCustomerMap, err := GetExistCustomerFromGorse()
	if err != nil {
		return err
	}

	subscriber, err := GetSubscriber()
	if err != nil {
		return err
	}

	if g.SubmitTask == nil {
		return errors.New("no callback function is set")
	}

	g.SubmitTask(func(group *sync.WaitGroup) {

		defer group.Done()
		var insertData []client.User
		//Find users who have not uploaded
		d3 := difference(customerId, existCustomer)
		//If the array is empty，return, the error message
		if len(d3) <= 0 {
			return
		}

		for _, s := range d3 {

			if len(s) <= 0 {
				continue
			}
			if val, exist := customerMap[s]; !exist {
				continue
			} else {
				var subscribe []string
				//if exist
				if s1, ex := subscriber[val.Id]; ex {
					subscribe = append(subscribe, s1...)

				}
				//insert data
				insertData = append(insertData, client.User{
					UserId:    Define.MakeItemId("customer", val.Id),
					Labels:    nil,
					Subscribe: subscribe,
					Comment:   val.Username,
				})
			}
		}

		if len(insertData) <= 0 {
			return
		}

		var num, limit = len(insertData), 1000
		var p, p1 = num / limit, num % limit
		if p1 != 0 {
			p = p + 1
		}

		for i := 0; i < p; i++ {

			var start, end = i * limit, (i + 1) * limit
			if start > num {
				break
			}
			if end > num {
				end = num
			}
			var data = insertData[start:end]

			g.SubmitTask(func(group *sync.WaitGroup) {
				err = addUser2Gorse(data)
				if err != nil {
					//record the error
					controllers.Log("insert user to gorse failed", err.Error(), controllers.LogError)
					return
				}
				defer group.Done()
			})
		}
	})

	g.SubmitTask(func(group *sync.WaitGroup) {

		defer group.Done()
		//Find the same elements
		se := SameElement(customerId, existCustomer)
		if len(se) <= 0 {
			return
		}
		for _, s := range se {

			if val, exist := existCustomerMap[s]; !exist {
				return
			} else {
				// Find existing subscribers for this user
				if subscribers, exist1 := subscriber[s]; exist1 {
					//Compare the recommendation system’s records with current data
					diffSubscriber := difference(subscribers, val)
					if len(diffSubscriber) <= 0 {
						continue
					}
					//Generate a new subscription list
					val = append(val, diffSubscriber...)
					//Get basic user information
					customer := customerMap[s]
					//generate new customer information
					c1 := client.UserPatch{
						Labels:    nil,
						Subscribe: val,
						Comment:   &customer.Username,
					}
					//Add a new thread to modify user information
					g.SubmitTask(func(group *sync.WaitGroup) {
						//done the thread
						defer group.Done()
						//Modify information
						err = editUser2Gorse(c1, Define.MakeItemId("customer", customer.Id))
						if err != nil {
							//record the error message
							controllers.Log("edit user to gorse", err.Error(), controllers.LogError)
							return
						}
					})
				}
			}
		}
	})
	return nil
}

func (g *Builder) BuildFeedback() error {

	feedbacks, err := GetFeedBack()
	if err != nil {
		return err
	}
	//Generate a temporary storage table with hash values as keys and inserted data as values
	var tmpMap = make(map[string]client.Feedback)
	var tmpHash []string

	for customerId, feedback := range feedbacks {
		for _, f := range feedback {
			hash := fmt.Sprintf(
				"%s-%s-%s",
				Define.MakeItemId("customer", customerId),
				Define.MakeItemId(f.ItemType, f.ItemId),
				f.Type)
			tmpHash = append(tmpHash, hash)
			tmpMap[hash] = client.Feedback{
				FeedbackType: f.Type,
				UserId:       Define.MakeItemId("customer", customerId),
				ItemId:       Define.MakeItemId(f.ItemType, f.ItemId),
				Timestamp:    time.Unix(f.Time, 0).Format("2006-01-02"),
			}
		}
	}

	feedback, err := GetExistFeedback(func(customerId, itemId, feedbackType string) string {
		return fmt.Sprintf(
			"%s-%s-%s",
			customerId,
			itemId,
			feedbackType)
	})
	if err != nil {
		return err
	}

	d1 := difference(tmpHash, feedback)
	if len(d1) <= 0 {
		return nil
	}

	var insertData []client.Feedback
	for _, s := range d1 {
		if data, exist := tmpMap[s]; !exist {
			continue
		} else {
			insertData = append(insertData, data)
		}
	}

	if len(insertData) <= 0 {
		return nil
	}

	var num, limit = len(insertData), 1000
	var p, p1 = num / limit, num % limit
	if p1 != 0 {
		p = p + 1
	}

	for i := 0; i < p; i++ {

		var start, end = i * limit, (i + 1) * limit
		if start > num {
			break
		}
		if end > num {
			end = num
		}
		var idata = insertData[start:end]

		g.SubmitTask(func(group *sync.WaitGroup) {
			//done the thread
			defer group.Done()
			//add to gorse
			if err := addFeedback2Gorse(idata); err != nil {
				//record the error
				controllers.Log("add feedback to gorse", err.Error(), controllers.LogError)
				return
			}
		})

	}
	return nil
}

func AddItemsToGorse(d []client.Item) error {

	var ctx = context.TODO()
	var cli = client.NewGorseClient(
		conf.AppConfig.Gorse.ToEndPoint(),
		conf.AppConfig.Gorse.ApiKey,
	)

	_, err := cli.InsertItems(ctx, d)
	if err != nil {
		return err
	}
	return nil
}

func send2MeiliSearch[dataModel any](indexName string, doc []dataModel) error {

	//与meiliSearch 连接
	cli := meilisearch.NewClient(meilisearch.ClientConfig{
		Host:   conf.AppConfig.MeiliSearch.ToHost(),
		APIKey: conf.AppConfig.MeiliSearch.ApiKey,
	})
	//add documents
	r1, err := cli.Index(indexName).
		AddDocuments(&doc, "id")
	if err != nil {
		//log
		controllers.Log("更新搜索索引失败", err.Error(), controllers.LogError)
		return err
	}
	if r1.Status != meilisearch.TaskStatusSucceeded && r1.Status != meilisearch.TaskStatusEnqueued {
		//return the error
		return errors.New(fmt.Sprintf("Return status exception,(%s)", r1.Status))
	}

	return nil
}

func addUser2Gorse(data []client.User) error {

	var ctx = context.TODO()
	var cli = client.NewGorseClient(
		conf.AppConfig.Gorse.ToEndPoint(),
		conf.AppConfig.Gorse.ApiKey,
	)

	_, err := cli.InsertUsers(ctx, data)
	if err != nil {
		return err
	}
	return nil

}

func editUser2Gorse(data client.UserPatch, userId string) error {

	var ctx = context.TODO()
	var cli = client.NewGorseClient(
		conf.AppConfig.Gorse.ToEndPoint(),
		conf.AppConfig.Gorse.ApiKey,
	)

	_, err := cli.UpdateUser(ctx, userId, data)
	if err != nil {
		return err
	}
	return nil
}

func addFeedback2Gorse(data []client.Feedback) error {

	var ctx = context.TODO()
	var cli = client.NewGorseClient(
		conf.AppConfig.Gorse.ToEndPoint(),
		conf.AppConfig.Gorse.ApiKey,
	)

	_, err := cli.InsertFeedback(ctx, data)
	if err != nil {
		return err
	}
	return nil
}
