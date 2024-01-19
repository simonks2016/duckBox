package Define

import (
	"errors"
	"regexp"
	"strings"
)

const (
	GorseCategoryVideo   = "video"
	GorseCategoryProgram = "program"
	GorseCategoryEpisode = "episode"
	GorseCategoryAd      = "ad"

	GorsePositiveFeedbackLike = "like"
	GorsePositiveFeedbackStar = "star"
	GorseReadFeedback         = "read"

	ClickEvent     = "click"
	ViewEvent      = "view"
	SubscribeEvent = "subscribe"
	PurchaseEvent  = "purchase"
	OrderEvent     = "order"
	BadEvent       = "bad"
	WatchEvent     = "watch"
)

func MakeItemId(t, id string) string {
	return strings.ToLower(t) + "-" + id
}

func SplitItemId(itemId string) ([]string, error) {

	p := `[a-zA-Z0-9]+`
	s1 := regexp.MustCompile(p).FindAllString(itemId, -1)
	//if the length is lower than 2
	if len(s1) < 2 {
		return nil, errors.New("an error occurred while retrieving the item ID")
	}
	return s1, nil
}

func StandardizedFeedbackEvents(s string) string {

	switch strings.ToLower(s) {
	case "like", "rating":
		return GorsePositiveFeedbackLike
	case ViewEvent, ClickEvent, WatchEvent:
		return GorseReadFeedback
	case SubscribeEvent, OrderEvent, PurchaseEvent:
		return GorsePositiveFeedbackStar
	default:
		return GorseReadFeedback
	}

}
