package main

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"github.com/ChimeraCoder/anaconda"
	"net/url"
	"os"
	"strconv"
)

// TwAPI has Api object and target user
type TwAPI struct {
	API  anaconda.TwitterApi
	User string
}

// Friends is slice of User
type Friends []anaconda.User

// FriendCursor has Friends and next cursor
type FriendCursor struct {
	Friends       Friends
	NextCursorStr string
}

// Ids is slice of id
type Ids []int64

// IDCursor has Ids and next cursor
type IDCursor struct {
	ids           Ids
	NextCursorStr string
}

// FriendsMap is map. Key is id, value is User
type FriendsMap map[int64]anaconda.User

// Following has following screenName and follower screenName
type Following struct {
	following string
	follower  string
}

// FollowingList is slice of Following
type FollowingList []Following

// Friends returns friend list by specified cursor
func (twa TwAPI) Friends(cursor string) FriendCursor {
	v := url.Values{"screen_name": {twa.User}, "cursor": {cursor}, "count": {"200"}}
	c, err := twa.API.GetFriendsList(v)
	if err != nil {
		panic(err)
	}
	return FriendCursor{c.Users, c.Next_cursor_str}
}

// AllFriends returns all friend list
func (twa TwAPI) AllFriends() Friends {
	friends := Friends{}

	for next := "-1"; ; {
		fc := twa.Friends(next)
		next = fc.NextCursorStr
		friends = append(friends, fc.Friends...)
		if next == "0" {
			break
		}
	}

	return friends
}

// FriendIds returns friend id list by specified cursor
func (twa TwAPI) FriendIds(cursor string) (IDCursor, error) {
	v := url.Values{"screen_name": {twa.User}, "cursor": {cursor}, "count": {"5000"}}
	c, err := twa.API.GetFriendsIds(v)
	return IDCursor{c.Ids, c.Next_cursor_str}, err
}

// AllFriendIds returns all friend id list
func (twa TwAPI) AllFriendIds() Ids {
	friendIds := Ids{}

	for next := "-1"; ; {
		ic, err := twa.FriendIds(next)
		if apiErr, ok := err.(*anaconda.ApiError); ok {
			// "error":"Not authorized."
			if apiErr.StatusCode == 401 {
				fmt.Println(fmt.Sprintf("%sは非公開です", twa.User))
				return Ids{}
			}
		}
		if err != nil {
			panic(err)
		}
		next = ic.NextCursorStr
		friendIds = append(friendIds, ic.ids...)
		if next == "0" {
			break
		}
	}

	return friendIds
}

// NewFriendsMap returns FriendsMap from Friends
func NewFriendsMap(friends Friends) FriendsMap {
	fmap := make(FriendsMap)
	for _, friend := range friends {
		fmap[friend.Id] = friend
	}
	return fmap
}

// SaveFriends saves friend list to csv
func SaveFriends(userName string, friends Friends) error {
	fout, err := os.OpenFile(userName+".csv", os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	writer := csv.NewWriter(fout)
	defer fout.Close()

	headers := []string{
		"Id",
		"ScreenName",
		"URL",
		"Name",
		"Description",
		"FriendsCount",
		"FollowersCount",
		"StatusesCount",
		"ListedCount",
		"Protected",
		"Lang",
		"Location",
		"TimeZone",
		"CreatedAt",
	}
	writer.Write(headers)

	for _, f := range friends {
		record := []string{
			f.IdStr,
			f.ScreenName,
			f.URL,
			f.Name,
			f.Description,
			strconv.Itoa(f.FriendsCount),
			strconv.Itoa(f.FollowersCount),
			strconv.FormatInt(f.StatusesCount, 10),
			strconv.FormatInt(f.ListedCount, 10),
			strconv.FormatBool(f.Protected),
			f.Lang,
			f.Location,
			f.TimeZone,
			f.CreatedAt,
		}
		writer.Write(record)
	}
	writer.Flush()
	return err
}

// SpecifiedFriends returns friends only in FriendsMap and Ids
func SpecifiedFriends(fmap FriendsMap, ids Ids) Friends {
	friends := Friends{}
	for _, id := range ids {
		if friend, ok := fmap[id]; ok {
			friends = append(friends, friend)
		}
	}
	return friends
}

// NewFollowingList returns FollowingList by following and followers
func NewFollowingList(following string, followers Friends) FollowingList {
	fList := FollowingList{}
	for _, follower := range followers {
		fList = append(fList, Following{following, follower.ScreenName})
	}
	return fList
}

// SaveFollowingList saves following list to csv
func SaveFollowingList(userName string, fList FollowingList) error {
	fout, err := os.OpenFile(userName+"_following_list.csv", os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	writer := csv.NewWriter(fout)
	defer fout.Close()

	headers := []string{
		"Following",
		"Follower",
	}
	writer.Write(headers)

	for _, following := range fList {
		record := []string{
			following.following,
			following.follower,
		}
		writer.Write(record)
	}
	writer.Flush()

	return err
}

type apiKey struct {
	ConsumerKey    string
	ConsumerSecret string
	AccessToken    string
	AccessSecret   string
}

func loadAPIKey() apiKey {
	file, err := os.Open("APIKEY.txt")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lines := []string{}
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return apiKey{
		lines[0],
		lines[1],
		lines[2],
		lines[3],
	}
}

func main() {
	apiKey := loadAPIKey()
	anaconda.SetConsumerKey(apiKey.ConsumerKey)
	anaconda.SetConsumerSecret(apiKey.ConsumerSecret)
	api := anaconda.NewTwitterApi(apiKey.AccessToken, apiKey.AccessSecret)
	//api.ReturnRateLimitError(true) // TODO: 消す

	userName := os.Args[1]
	twAPI := TwAPI{*api, userName}
	friends := twAPI.AllFriends()
	friendsMap := NewFriendsMap(friends)
	fmt.Println(fmt.Sprintf("%sのフォロー数は%d人", twAPI.User, len(friends)))
	err := SaveFriends(twAPI.User, friends)
	if err != nil {
		panic(err)
	}

	followingList := FollowingList{}
	for _, friend := range friends {
		fmt.Println(friend.ScreenName)
		friendAPI := TwAPI{*api, friend.ScreenName}
		ids := friendAPI.AllFriendIds()
		followers := SpecifiedFriends(friendsMap, ids)
		mutualList := NewFollowingList(friendAPI.User, followers)
		fmt.Println(fmt.Sprintf("%sと%sが共通にフォローしている人数は%d人", twAPI.User, friendAPI.User, len(mutualList)))
		followingList = append(followingList, mutualList...)
	}
	err = SaveFollowingList(twAPI.User, followingList)
	if err != nil {
		panic(err)
	}
}
