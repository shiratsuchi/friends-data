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

type TwApi struct {
	Api  anaconda.TwitterApi
	User string
}

type Friends []anaconda.User

type FriendCursor struct {
	Friends         Friends
	Next_cursor_str string
}

type Ids []int64

type IdCursor struct {
	ids             Ids
	Next_cursor_str string
}

type FriendsMap map[int64]anaconda.User

type Following struct {
	following string
	follower  string
}

type FollowingList []Following

func (twa TwApi) Friends(next_cursor string) FriendCursor {
	v := url.Values{"screen_name": {twa.User}, "cursor": {next_cursor}, "count": {"200"}}
	cursor, err := twa.Api.GetFriendsList(v)
	if err != nil {
		panic(err)
	}
	return FriendCursor{cursor.Users, cursor.Next_cursor_str}
}

func (twa TwApi) AllFriends() Friends {
	friends := Friends{}

	for next := "-1"; ; {
		fc := twa.Friends(next)
		next = fc.Next_cursor_str
		friends = append(friends, fc.Friends...)
		if next == "0" {
			break
		}
	}

	return friends
}

func (twa TwApi) FriendIds(next_cursor string) (IdCursor, error) {
	v := url.Values{"screen_name": {twa.User}, "cursor": {next_cursor}, "count": {"5000"}}
	cursor, err := twa.Api.GetFriendsIds(v)
	return IdCursor{cursor.Ids, cursor.Next_cursor_str}, err
}

func (twa TwApi) AllFriendIds() Ids {
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
		next = ic.Next_cursor_str
		friendIds = append(friendIds, ic.ids...)
		if next == "0" {
			break
		}
	}

	return friendIds
}

func MakeFriendsMap(friends Friends) FriendsMap {
	fmap := make(FriendsMap)
	for _, friend := range friends {
		fmap[friend.Id] = friend
	}
	return fmap
}

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

func MutualFriends(fmap FriendsMap, ids Ids) Friends {
	friends := Friends{}
	for _, id := range ids {
		if friend, ok := fmap[id]; ok {
			friends = append(friends, friend)
		}
	}
	return friends
}

func CreateMutualList(following string, friendsMap FriendsMap, ids Ids) FollowingList {
	followers := MutualFriends(friendsMap, ids)

	fList := FollowingList{}
	for _, follower := range followers {
		fList = append(fList, Following{following, follower.ScreenName})
	}
	return fList
}

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

type ApiKey struct {
	ConsumerKey    string
	ConsumerSecret string
	AccessToken    string
	AccessSecret   string
}

func LoadApiKey() ApiKey {
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
	return ApiKey{
		lines[0],
		lines[1],
		lines[2],
		lines[3],
	}
}

func main() {
	apiKey := LoadApiKey()
	anaconda.SetConsumerKey(apiKey.ConsumerKey)
	anaconda.SetConsumerSecret(apiKey.ConsumerSecret)
	api := anaconda.NewTwitterApi(apiKey.AccessToken, apiKey.AccessSecret)
	//api.ReturnRateLimitError(true) // TODO: 消す

	userName := os.Args[1]
	twApi := TwApi{*api, userName}
	friends := twApi.AllFriends()
	friendsMap := MakeFriendsMap(friends)
	fmt.Println(fmt.Sprintf("%sのフォロー数は%d人", twApi.User, len(friends)))
	err := SaveFriends(twApi.User, friends)
	if err != nil {
		panic(err)
	}

	followingList := FollowingList{}
	for _, friend := range friends {
		fmt.Println(friend.ScreenName)
		friendApi := TwApi{*api, friend.ScreenName}
		ids := friendApi.AllFriendIds()
		mutualList := CreateMutualList(friendApi.User, friendsMap, ids)
		fmt.Println(fmt.Sprintf("%sと%sが共通にフォローしている人数は%d人", twApi.User, friendApi.User, len(mutualList)))
		followingList = append(followingList, mutualList...)
	}
	err = SaveFollowingList(twApi.User, followingList)
	if err != nil {
		panic(err)
	}
}
