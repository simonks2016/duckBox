package DataModel

import (
	"DuckBox/auth"
	"encoding/json"
	"fmt"
	"github.com/astaxie/beego/orm"
	"sort"
	"strings"
	"time"
)

const (
	BenefitsOpenProgramPromotion = "OpenProgramPromotion" //打开节目推广
	BenefitsPriorityDisplay      = "PriorityDisplay"      //优先显示

	BenefitsAccessToPaidPrograms         = "AccessToPaidPrograms"         //允许开通收费节目
	BenefitsAccessToMerchants            = "AccessToMerchants"            //允许开通商家频道
	BenefitsAccessToLive                 = "AccessToLive"                 //允许直播
	BenefitsAccessToPublishLongTimeVideo = "AccessToPublishLongTimeVideo" //允许发布长视频
	BenefitsAccessToCreateFansGroup      = "AccessToFansGroup"            //允许创建新粉丝群组
	BenefitsPermissionToUseAI            = "PermissionToUseAI"            //允许试用AI
	BenefitsKOLExclusiveLogo             = "KOLExclusiveLogo"             //网红专属标识
	BenefitsMerchantExclusiveLogo        = "MerchantExclusiveLogo"        //商家专属标识
	BenefitsMerchantExclusiveAds         = "MerchantExclusiveAds"         //商家专属广告
	BenefitsExclusiveChannelKeywords     = "ExclusiveChannelKeywords"     //专属频道关键

	BenefitsVideoCDNAcceleration      = "CDNAcceleration"        //CDN加速
	BenefitsVideoDownload             = "VideoDownload"          //视频下载
	BenefitsUltraClear                = "UltraClear"             //超高清
	BenefitsNoAds                     = "NoAds"                  //无广告干扰
	BenefitsCastScreen                = "CastScreen"             //投屏
	BenefitsMemberExclusiveContent    = "MemberExclusiveContent" //会员专属内容
	BenefitsShareMemberShipWithFamily = "ShareWithFamilyMembers" //分享给家庭成员
	BenefitsShareMemberShipWithTeam   = "ShareWithGroupMembers"

	BenefitsMultipleAdSlotsPromotingItem = "MultipleAdSlotsPromotingItem" //允许一个节目创建多个广告

	CurrencyUSDollar = "USD"
	CurrencyRMB      = "RMB"
)

type SubscribeMember struct {
	Id         string `orm:"pk"`
	DeadLine   int64
	CreateTime int64
	Signature  string
	Benefits   string
	Customer   *Customer `orm:"rel(fk);null"`
	Order      *Order    `orm:"rel(fk);null"`
}

type MemberPackage struct {
	Id                   string `orm:"pk"`
	Name                 string
	Description          string
	ExpireTime           int64
	State                int
	Benefits             string
	Price                float64
	OriginalPrice        float64
	Currency             string
	WhetherContinuous    bool      //是否连续包月
	EnterpriseLevel      bool      //是否企业级
	WhetherShareMember   bool      //是否允许分享会籍
	ShareMemberMaxNumber int       //最大分享会籍数量
	Applicant            *Employee `orm:"rel(fk)"`
}

func DisplayBenefits() map[string]string {

	var b = make(map[string]string)

	b[BenefitsAccessToMerchants] = "允许开通商家频道"
	b[BenefitsAccessToLive] = "允许开通直播频道"
	b[BenefitsCastScreen] = "允许投屏"
	b[BenefitsAccessToCreateFansGroup] = "允许创建粉丝群组"
	b[BenefitsAccessToPaidPrograms] = "允许创建收费节目"
	b[BenefitsNoAds] = "无广告干扰"
	b[BenefitsAccessToPublishLongTimeVideo] = "允许发布长视频"
	b[BenefitsExclusiveChannelKeywords] = "频道专属关键词"
	b[BenefitsKOLExclusiveLogo] = "达人专属标识"
	b[BenefitsMemberExclusiveContent] = "会员专属内容"
	b[BenefitsMerchantExclusiveAds] = "商家专属广告"
	b[BenefitsMerchantExclusiveLogo] = "商家专属标识"
	b[BenefitsMultipleAdSlotsPromotingItem] = "允许一个节目创建多个广告"
	b[BenefitsOpenProgramPromotion] = "打开节目推广工具"
	b[BenefitsPermissionToUseAI] = "允许试用AI"
	b[BenefitsPriorityDisplay] = "优先显示"
	b[BenefitsShareMemberShipWithFamily] = "分享给家庭成员"
	b[BenefitsUltraClear] = "超高清画质"
	b[BenefitsVideoCDNAcceleration] = "CDN加速线路"
	b[BenefitsVideoDownload] = "视频可下载"
	b[BenefitsShareMemberShipWithTeam] = "分享给团队成员"

	return b
}

func (s *SubscribeMember) VerifySignature() bool {
	return strings.Compare(s.Signature, s.MakeSignature()) == 0
}
func (s *SubscribeMember) MakeSignature() string {

	var source []string

	source = append(source, fmt.Sprintf("id=%s", s.Id),
		fmt.Sprintf("deadline=%d", int(s.DeadLine)),
		fmt.Sprintf("create_time=%d", int(s.CreateTime)),
		fmt.Sprintf("customer_id=%s", s.Customer.Id),
	)

	var b = make(map[string]bool)
	var ben []string

	//Unmarshal JSON
	err := json.Unmarshal([]byte(s.Benefits), &b)
	if err != nil {
		return ""
	}
	//make benefits slice
	for k, v := range b {
		if v == true {
			ben = append(ben, k)
		}
	}
	//sort benefits slice
	sort.Strings(ben)
	//make md5 string
	m5 := auth.M5(strings.Join(ben, "&&"))
	//push into slice
	source = append(source, fmt.Sprintf("benefits=%s", m5))
	//sort
	sort.Strings(source)
	//to md5
	return auth.HMAC(auth.M5(strings.Join(source, "&&")), "yimi.tv-member-key")
}

func (c *Customer) IsSubscribeMember() bool {

	var (
		o = orm.NewOrm()
		s SubscribeMember
	)
	//查阅是否有记录
	if err := o.QueryTable(&SubscribeMember{}).Filter("Customer", c.Id).One(&s); err != nil {
		return false
	}
	//检查签名和有效期
	return s.VerifySignature() && s.DeadLine > time.Now().Unix()
}

func (this *SubscribeMember) HaveThisBenefits(benefits string) bool {
	var ben = this.Benefits
	var benefitsMap = make(map[string]bool)
	if len(ben) <= 0 {
		return false
	}

	if err := json.Unmarshal([]byte(ben), &benefitsMap); err != nil {
		fmt.Println(err.Error())
		return false
	}

	if val, exist := benefitsMap[benefits]; exist == false {
		return false
	} else {
		return val
	}
}
func (this *MemberPackage) ToBenefitsContent() (error, []string) {
	var d = make(map[string]bool)
	var result []string

	if err := json.Unmarshal([]byte(this.Benefits), &d); err != nil {
		return err, nil
	}
	for k, v := range d {
		if v == true {
			result = append(result, k)
		}
	}
	return nil, result
}
