package Rebuilder

import (
	"DuckBox/DataModel"
	"DuckBox/Define"
	"DuckBox/conf"
	"context"
	"errors"
	"fmt"
	"github.com/astaxie/beego/orm"
	"github.com/meilisearch/meilisearch-go"
	"github.com/zhenghaoz/gorse/client"
	"strings"
	"time"
)

func GetVideosFromMySQL() (map[string]*DataModel.Video, []string, error) {

	var (
		o           = orm.NewOrm()
		result      []*DataModel.Video
		responseIds []string
		response    = make(map[string]*DataModel.Video)
	)

	//get total
	if num, err := o.QueryTable(&DataModel.Video{}).Filter("State", DataModel.VideoStatusNormal).Count(); err != nil {
		return nil, nil, err
	} else {

		var p, p1 = num / 1000, num % 1000
		if p1 != 0 {
			p = p + 1
		}

		for i := 0; i < int(p); i++ {
			//calculate start point
			var start = i * 1000
			//loop query doc
			if _, err := o.QueryTable(&DataModel.Video{}).
				Filter("State", DataModel.VideoStatusNormal).
				Offset(start).Limit(1000).
				All(&result); err != nil {
				//if you get an error message
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
		}
	}
	return response, responseIds, nil
}

func GetProgramsFromMySQL() (map[string]*DataModel.Program, []string, error) {

	var (
		o           = orm.NewOrm()
		result      []*DataModel.Program
		responseIds []string
		response    = make(map[string]*DataModel.Program)
	)

	if num, err := o.QueryTable(&DataModel.Program{}).Filter("State", 1).Count(); err != nil {
		return nil, nil, err
	} else {

		var p, p1 = num / 1000, num % 1000
		if p1 != 0 {
			p = p + 1
		}
		for i := 0; i < int(p); i++ {
			//cal
			var start = i * 1000
			//loop to query document
			if _, err := o.QueryTable(&DataModel.Program{}).
				Filter("State", 1).Offset(start).Limit(1000).All(&result); err != nil {
				//if you get the error message
				return nil, nil, err
			}
			//get ids
			for _, program := range result {
				//
				responseIds = append(responseIds, program.Id)
				//response
				if _, exist := response[program.Id]; !exist {
					response[program.Id] = program
				}
			}
		}
	}

	return response, responseIds, nil
}

func GetExistingDocumentsID(indexName string) ([]string, error) {

	client := meilisearch.NewClient(
		meilisearch.ClientConfig{
			Host:   conf.AppConfig.MeiliSearch.ToHost(),
			APIKey: conf.AppConfig.MeiliSearch.ApiKey,
		},
	)

	isExistIndex := CheckIndexExist(indexName, client)

	//if not exist ,then create index in search
	if isExistIndex == false {
		//create index
		_, err := client.CreateIndex(&meilisearch.IndexConfig{
			Uid:        indexName,
			PrimaryKey: "id",
		})
		if err != nil {
			return nil, err
		}
		//try 3 times
		for i := 0; i < 3; i++ {
			//sleep the 2 minute
			time.Sleep(time.Minute * 2)
			//check index is exist
			isExistIndex = CheckIndexExist(indexName, client)
			//if index is exist
			if isExistIndex {
				break
			}
		}
		if !isExistIndex {
			return nil, errors.New("the create index is failed")
		}
	}

	var result meilisearch.DocumentsResult
	var query meilisearch.DocumentsQuery
	var response []string

	query.Limit = 200
	query.Offset = 0
	query.Fields = []string{"id"}

	//get all document from search
	err := client.Index(indexName).GetDocuments(&query, &result)
	if err != nil {
		//log
		return nil, err
	}

	//loop to get result ,append the list
	for _, m := range result.Results {
		response = append(response, fmt.Sprintf("%v", m["id"]))
	}

	//If the returned quantity is greater than the limit quantity
	if result.Total > query.Limit {
		//获取页数
		var page = int(result.Total / query.Limit)
		//if
		if result.Total%query.Limit != 0 {
			page = page + 1
		}

		//loop the to get result
		for i := 0; i < page; i++ {
			//loop to get every page
			err = client.Index(indexName).GetDocuments(&meilisearch.DocumentsQuery{
				Offset: int64((page - 1) * 200),
				Limit:  query.Limit,
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

func difference(newSlice []string, existSlice []string) (response []string) {

	var a = make(map[string]bool)
	for _, s := range existSlice {
		//if exist the
		if val, exist := a[s]; exist {
			continue
		} else if !val {
			a[s] = true
		} else {
			a[s] = true
		}
	}

	for _, s := range newSlice {
		//if they in the d2
		if _, exist := a[s]; !exist {
			//append the response
			response = append(response, s)
		}
	}
	return response
}

func SameElement(d []string, existElement []string) (response []string) {

	var a = make(map[string]bool)

	for _, s := range existElement {
		//if exist the
		if val, exist := a[s]; exist {
			continue
		} else if !val {
			a[s] = true
		} else {
			a[s] = true
		}
	}

	for _, s := range d {
		//if they in the d2
		if _, exist := a[s]; exist {
			//append the response
			response = append(response, s)
		}
	}
	return response
}

func GetExistItemsId() (map[string][]string, error) {

	var cli = client.NewGorseClient(
		conf.AppConfig.Gorse.ToEndPoint(),
		conf.AppConfig.Gorse.ApiKey,
	)
	var existDocumentIds []string
	var eMap = make(map[string][]string)
	var ctx = context.TODO()

	err, s := getDataFromGorseClient("", 100, &existDocumentIds, cli, ctx)
	if err != nil {
		return nil, err
	}

	for {
		if len(s) <= 0 {
			break
		}
		err, s = getDataFromGorseClient(s, 100, &existDocumentIds, cli, ctx)
		if err != nil {
			return nil, err
		}
	}

	for _, id := range existDocumentIds {

		itemId, err := Define.SplitItemId(id)
		if err != nil {
			return nil, err
		}

		if len(itemId) < 2 {
			continue
		}

		key := strings.ToLower(itemId[0])

		if val, exist := eMap[key]; !exist {
			eMap[key] = []string{itemId[1]}
		} else {
			//append
			val = append(val, itemId[1])
			//copy
			eMap[key] = val
		}
	}

	return eMap, nil
}

func getDataFromGorseClient(cursor string, num int, p *[]string, cli *client.GorseClient, ctx context.Context) (error, string) {

	//get items
	items, err := cli.GetItems(ctx, cursor, num)
	if err != nil {
		return err, ""
	}
	for _, item := range items.Items {
		*p = append(*p, item.ItemId)
	}
	return nil, items.Cursor
}

func GetEpisodeMap() (map[string]bool, error) {

	var response = make(map[string]bool)
	var o = orm.NewOrm()
	var e []*DataModel.Episodes

	if num, err := o.QueryTable(&DataModel.Episodes{}).Count(); err != nil {
		return nil, err
	} else {

		var p, p1 = num / 1000, num % 1000
		if p1 != 0 {
			p = p + 1
		}
		for i := 0; i < int(p); i++ {
			//calculate the start point
			var start = i * 1000
			//loop get the episodes
			if _, err := o.QueryTable(&DataModel.Episodes{}).Offset(start).
				Limit(1000).All(&e); err != nil {
				return nil, err
			}
			//loop the set video id to map key
			for _, episodes := range e {
				if episodes.Video == nil {
					if _, err = o.LoadRelated(episodes, "Video"); err != nil {
						if errors.Is(err, orm.ErrNoRows) {
							continue
						}
						return nil, err
					}
				}
				response[episodes.Video.Id] = true
			}
		}

	}

	return response, nil
}

func CheckIndexExist(indexName string, client *meilisearch.Client) bool {

	var isExistIndex bool

	indexes, err := client.GetIndexes(&meilisearch.IndexesQuery{
		Limit:  20,
		Offset: 0,
	})
	if err != nil {
		return false
	}
	//Loop to get the result, if it exists
	for _, result := range indexes.Results {

		if strings.Compare(result.UID, indexName) == 0 {
			isExistIndex = true
		}
	}
	return isExistIndex
}

func GetExistCustomerIdFromDB() ([]string, map[string]*DataModel.Customer, error) {

	var o = orm.NewOrm()
	var customers []*DataModel.Customer
	var responseIDs []string
	var responseMap = make(map[string]*DataModel.Customer)

	if num, err := o.QueryTable(&DataModel.Customer{}).Count(); err != nil {
		return nil, nil, err
	} else {

		var p, p1 = num / 1000, num % 1000
		if p1 != 0 {
			p = p + 1
		}

		for i := 0; i < int(p); i++ {
			var start = i * 1000

			if _, err = o.QueryTable(&DataModel.Customer{}).Offset(start).
				Limit(1000).All(&customers); err != nil {

				return nil, nil, err
			}

			for _, customer := range customers {
				if customer.State != DataModel.VideoStatusNormal {
					continue
				}
				if _, exist := responseMap[customer.Id]; !exist {
					//append the ids
					responseIDs = append(responseIDs, customer.Id)
					//add the map
					responseMap[customer.Id] = customer
				} else {
					continue
				}
			}
		}
	}
	return responseIDs, responseMap, nil
}

func GetSubscriber() (map[string][]string, error) {
	var o = orm.NewOrm()
	var result []orm.Params
	var responseMap = make(map[string][]string)

	if num, err := o.QueryTable(&DataModel.Follow{}).Count(); err != nil {
		return nil, err
	} else {

		var p, p1 = num / 1000, num % 1000
		if p1 != 0 {
			p = p + 1
		}

		for i := 0; i < int(p); i++ {
			var start = i * 1000

			if _, err = o.QueryTable(&DataModel.Follow{}).Offset(start).
				Limit(1000).Values(&result, "followers_id", "leader_id"); err != nil {

				return nil, err
			}

			for _, params := range result {
				leaderId := fmt.Sprintf("%v", params["leader_id"])
				followersId := fmt.Sprintf("%v", params["followers_id"])

				if val, exist := responseMap[leaderId]; !exist {
					responseMap[leaderId] = []string{followersId}
				} else {
					//append to value
					val = append(val, followersId)
					responseMap[leaderId] = val
				}
			}
		}
	}
	return responseMap, nil
}

type FeedBackOfDB struct {
	Type     string `json:"type"`
	Time     int64  `json:"time"`
	ItemId   string `json:"item_id"`
	ItemType string `json:"item_type"`
}

func GetLike() (map[string][]*FeedBackOfDB, error) {

	var o = orm.NewOrm()
	var response = make(map[string][]*FeedBackOfDB)

	if num, err := o.QueryTable(&DataModel.Like{}).Count(); err != nil {
		return nil, err
	} else {

		var p, p1 = num / 1000, num % 1000
		if p1 != 0 {
			p = p + 1
		}

		//create function of check item type
		checkItemType := func(s string) string {
			if len(s) <= 0 {
				return "video"
			}
			return "program"
		}
		//create function of get item id
		getItemId := func(s, s1 string) string {
			if len(s) <= 0 {
				return s1
			}
			return s
		}
		//create function of set feedback type
		sft := func(t int) string {
			if t == 1 {
				return Define.GorsePositiveFeedbackLike
			}
			return ""
		}

		for i := 0; i < int(p); i++ {
			var par []*DataModel.Like
			if _, err = o.QueryTable(&DataModel.Like{}).
				All(&par, "type", "applicant_id", "video_id", "topic_id", "create_time"); err != nil {
				return nil, err
			}

			for _, params := range par {

				applicantId := params.Applicant.Id
				createTime := params.CreateTime
				var videoId, programId string
				//
				if params.Video != nil {
					videoId = params.Video.Id
				}
				if params.Topic != nil {
					programId = params.Topic.Id
				}
				t := params.Type

				if val, exist := response[applicantId]; !exist {
					response[applicantId] = []*FeedBackOfDB{
						{
							Type:     sft(t),
							Time:     createTime,
							ItemId:   getItemId(videoId, programId),
							ItemType: checkItemType(programId),
						},
					}
				} else {

					val = append(val, &FeedBackOfDB{
						Type:     sft(t),
						Time:     createTime,
						ItemId:   getItemId(videoId, programId),
						ItemType: checkItemType(programId),
					})
					response[applicantId] = val
				}
			}
		}
	}

	return response, nil
}

func GetViewRecord() (map[string][]*FeedBackOfDB, error) {

	var o = orm.NewOrm()
	var response = make(map[string][]*FeedBackOfDB)

	if num, err := o.QueryTable(&DataModel.Record{}).
		Filter("ItemType__in", "video", "program").Count(); err != nil {
		return nil, err
	} else {

		var p, p1 = num / 1000, num % 1000
		if p1 != 0 {
			p = p + 1
		}

		for i := 0; i < int(p); i++ {
			var par []*DataModel.Record
			if _, err = o.QueryTable(&DataModel.Record{}).Filter("ItemType__in", "video", "program").
				All(&par); err != nil {
				return nil, err
			}

			for _, params := range par {

				customerId := params.CustomerId
				itemId := params.ItemId
				itemType := params.ItemType
				event := params.Event
				happenTime := params.HappenTime

				if len(customerId) <= 0 {
					continue
				}
				if val, exist := response[customerId]; !exist {
					response[customerId] = []*FeedBackOfDB{
						{
							Type:     Define.StandardizedFeedbackEvents(event),
							Time:     happenTime,
							ItemId:   itemId,
							ItemType: strings.ToLower(itemType),
						},
					}
				} else {

					val = append(val, &FeedBackOfDB{
						Type:     Define.StandardizedFeedbackEvents(event),
						Time:     happenTime,
						ItemId:   itemId,
						ItemType: strings.ToLower(itemType),
					})
					response[customerId] = val
				}
			}
		}
	}

	return response, nil
}

func GetFeedBack() (map[string][]*FeedBackOfDB, error) {

	like, err := GetLike()
	if err != nil {
		return nil, err
	}

	record, err := GetViewRecord()
	if err != nil {
		return nil, err
	}

	var response = make(map[string][]*FeedBackOfDB)

	for key, dbs := range like {

		if val, exist := response[key]; !exist {
			//if you have the customer id not in map,then new map
			response[key] = dbs
		} else {
			//copy
			val = append(val, dbs...)
			//make
			response[key] = val
		}
	}

	for key, dbs := range record {

		if val, exist := response[key]; !exist {
			//if you have the customer id not in map,then new map
			response[key] = dbs
		} else {
			//copy
			val = append(val, dbs...)
			//make
			response[key] = val
		}
	}

	return response, nil
}

func GetExistCustomerFromGorse() ([]string, map[string][]string, error) {

	cli := client.NewGorseClient(
		conf.AppConfig.Gorse.ToEndPoint(),
		conf.AppConfig.Gorse.ApiKey,
	)
	var ctx = context.TODO()
	var response []string
	var subscribeMap = make(map[string][]string)

	appendResponse := func(d []client.User, resp *[]string, sm *map[string][]string) {

		ism := *sm
		for _, user := range d {
			id, err := Define.SplitItemId(user.UserId)
			if err != nil {
				return
			}
			//
			customerId := id[1]
			*resp = append(*resp, customerId)
			//subscriber
			var subscriber []string
			//get subscriber
			for _, s := range user.Subscribe {
				id1, err := Define.SplitItemId(s)
				if err != nil {
					return
				}
				subscriber = append(subscriber, id1[1])
			}

			if val, exist := ism[customerId]; !exist {
				ism[customerId] = subscriber
			} else {
				//add the old list
				val = append(val, subscriber...)
				//copy the new value
				ism[customerId] = val
			}
		}
	}

	users, err := cli.GetUsers(ctx, "", 100)
	if err != nil {
		return nil, nil, err
	}
	appendResponse(users.Users, &response, &subscribeMap)

	if len(users.Cursor) > 0 {
		//cursor
		cursor := users.Cursor
		for {
			if len(cursor) <= 0 {
				break
			}
			users, err = cli.GetUsers(ctx, users.Cursor, 100)
			if err != nil {
				return nil, nil, err
			}
			cursor = users.Cursor
			appendResponse(users.Users, &response, &subscribeMap)
		}
	}
	return response, subscribeMap, nil
}

func GetExistFeedback(handler func(customerId, itemId, feedbackType string) string) ([]string, error) {

	cli := client.NewGorseClient(
		conf.AppConfig.Gorse.ToEndPoint(),
		conf.AppConfig.Gorse.ApiKey,
	)
	var ctx = context.TODO()
	var response []string

	feedback, err := cli.GetFeedback(ctx, "", 100)
	if err != nil {
		return nil, err
	}

	for _, f := range feedback.Feedback {

		//append
		response = append(response, handler(f.UserId, f.ItemId, f.FeedbackType))
	}
	return response, nil
}
