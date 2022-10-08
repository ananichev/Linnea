package models

import "encoding/json"

type UserRank string

const (
	UserRankUser  UserRank = "user"
	UserRankAdmin UserRank = "admin"
)

type User struct {
	// The users Twitch Name
	Name      string   `json:"name"`
	TwitchUID string   `json:"twitch_uid"`
	Rank      UserRank `json:"rank"`
}

func (u User) ToString() (string, error) {
	b, err := json.Marshal(u)
	return string(b), err
}

func UserFromTwitch(name, uid string) User {
	return User{
		Name:      name,
		TwitchUID: uid,
		Rank:      UserRankUser,
	}
}
