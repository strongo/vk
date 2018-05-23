package vk

import (
	"encoding/json"
	"net/http"
	"strings"
	"strconv"
	"fmt"
	"context"
	"io/ioutil"
	"github.com/strongo/log"
	"github.com/pkg/errors"
)

var (
	// NameCases is a list of name cases available for VK
	NameCases = []string{"nom", "gen", "dat", "acc", "ins", "abl"}
)

const (
	FieldFirstName = "first_name"
	FieldLastName = "last_name"
	FieldScreenName = "screen_name"
	FieldNickname = "nickname"
)

type (

	VkError interface {
		VkErrorCode() int
	}

	RequestParam struct {
		Key string `json:"key"`
		Value string `json:"value"`
	}

	vkError struct {
		Code    int `json:"error_code"`
		Message string `json:"error_msg"`
		RequestParams []RequestParam `json:"request_params"`
	}

	// Response from users.get
	Response struct {
		Error *vkError `json:"error"`
		Response []UserInfo `json:"response"`
	}

	// UserInfo contains user information
	// TODO improve fields list from here: http://vk.com/dev/fields
	UserInfo struct {
		ID         int          `json:"id"`
		FirstName  string       `json:"first_name"`
		LastName   string       `json:"last_name"`
		ScreenName string       `json:"screen_name"`
		Nickname   string       `json:"nickname"`
		Sex        int          `json:"sex,omitempty"`
		Domain     string       `json:"domain,omitempty"`
		Birthdate  string       `json:"bdate,omitempty"`
		City       GeoPlace     `json:"city,omitempty"`
		Country    GeoPlace     `json:"country,omitempty"`
		Photo50    string       `json:"photo_50,omitempty"`
		Photo100   string       `json:"photo_100,omitempty"`
		Photo200   string       `json:"photo_200,omitempty"`
		PhotoMax               string       `json:"photo_max,omitempty"`
		Photo200Orig           string       `json:"photo_200_orig,omitempty"`
		PhotoMaxOrig           string       `json:"photo_max_orig,omitempty"`
		HasMobile              bool         `json:"has_mobile,omitempty"`
		Online                 bool         `json:"online,omitempty"`
		CanPost                bool         `json:"can_post,omitempty"`
		CanSeeAllPosts         bool         `json:"can_see_all_posts,omitempty"`
		CanSeeAudio            bool         `json:"can_see_audio,omitempty"`
		CanWritePrivateMessage bool         `json:"can_write_private_message,omitempty"`
		Site                   string       `json:"site,omitempty"`
		Status                 string       `json:"status,omitempty"`
		LastSeen               PlatformInfo `json:"last_seen,omitempty"`
		CommonCount            int          `json:"common_count,omitempty"`
		University             int          `json:"university,omitempty"`
		UniversityName         string       `json:"university_name,omitempty"`
		Faculty                int          `json:"faculty,omitempty"`
		FacultyName            int          `json:"faculty_name,omitempty"`
		Graduation             int          `json:"graduation,omitempty"`
		Relation               int          `json:"relation,omitempty"`
		Universities           []University `json:"universities,omitempty"`
		Schools                []School     `json:"schools,omitempty"`
		Relatives              []Relative   `json:"relatives,omitempty"`
	}
	// GeoPlace contains geographical information like City, Country
	GeoPlace struct {
		ID    int    `json:"id"`
		Title string `json:"title"`
	}
	// PlatformInfo contains information about time and platform
	PlatformInfo struct {
		Time     EpochTime `json:"time"`
		Platform int       `json:"platform"`
	}
	// University contains information about the university
	University struct {
		ID              int    `json:"id"`
		Country         int    `json:"country"`
		City            int    `json:"city"`
		Name            string `json:"name"`
		Faculty         int    `json:"faculty"`
		FacultyName     string `json:"faculty_name"`
		Chair           int    `json:"chair"`
		ChairName       string `json:"chair_name"`
		Graduation      int    `json:"graduation"`
		EducationForm   string `json:"education_form"`
		EducationStatus string `json:"education_status"`
	}
	// School contains information about schools
	School struct {
		ID         int    `json:"id"`
		Country    int    `json:"country"`
		City       int    `json:"city"`
		Name       string `json:"name"`
		YearFrom   int    `json:"year_from"`
		YearTo     int    `json:"year_to"`
		Class      string `json:"class"`
		TypeStr    string `json:"type_str,omitempty"`
		Speciality string `json:"speciality,omitempty"`
	}
	// Relative contains information about relative to the user
	Relative struct {
		ID   int    `json:"id"`   // negative id describes non-existing users (possibly prepared id if they will register)
		Type string `json:"type"` // like `parent`, `grandparent`, `sibling`
		Name string `json:"name,omitempty"`
	}
)

func (err vkError) Error() string {
	return fmt.Sprintf("Code=%v, len(request_params)=%v, %v", err.Code, len(err.RequestParams), err.Message)
}

func (err vkError) VkErrorCode() int {
	return err.Code
}

func (api *API) GetUserByIntID(c context.Context, userID int64, nameCase string, fields ...string) (UserInfo, error) {
	if users, err := api.UsersGet(c, []string{strconv.FormatInt(userID, 10)}, fields, nameCase); err != nil {
		return UserInfo{}, err
	} else if len(users) != 1 {
			panic(fmt.Sprintf("len(users):%v != 1", len(users)))
	} else {
		return users[0], nil
	}
}

// UsersGet implements method http://vk.com/dev/users.get
//
//     userIds - no more than 1000, use `user_id` or `screen_name`
//     fields - sex, bdate, city, country, photo_50, photo_100, photo_200_orig,
//     photo_200, photo_400_orig, photo_max, photo_max_orig, online,
//     online_mobile, lists, domain, has_mobile, contacts, connections, site,
//     education, universities, schools, can_post, can_see_all_posts,
//     can_see_audio, can_write_private_message, status, last_seen,
//     common_count, relation, relatives, counters
//     name_case - choose one of nom, gen, dat, acc, ins, abl.
//     nom is default
//
func (api *API) UsersGet(c context.Context, userIds []string, fields []string, nameCase string) ([]UserInfo, error) {
	if len(userIds) == 0 {
		return nil, errors.New("you must pass at least one id or screen_name")
	}
	if !ElemInSlice(nameCase, NameCases) {
		return nil, errors.New("the only available name cases are: " + strings.Join(NameCases, ", "))
	}

	endpoint := api.getAPIURL("users.get")
	query := endpoint.Query()
	query.Set("user_ids", strings.Join(userIds, ","))

	if len(fields) > 0 {
		fieldsStr := strings.Join(fields, ",")
		log.Debugf(c, "VK fields: "+fieldsStr)
		query.Set("fields", fieldsStr)
	}
	if nameCase != "" {
		query.Set("name_case", nameCase)
	}

	endpoint.RawQuery = query.Encode()

	var err error
	var resp *http.Response
	var response Response

	httpClient := api.httpClient
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	url := endpoint.String()
	log.Debugf(c, "url: %v", url)
	if resp, err = httpClient.Get(url); err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	responseBody, err := ioutil.ReadAll(resp.Body)

	log.Debugf(c, "VK response(status=%v) body: %v", resp.StatusCode, string(responseBody))

	if err = json.Unmarshal(responseBody, &response); err != nil {
		return nil, errors.Wrap(err, "Failed to unmarshal VK response")
	}
	log.Debugf(c, "Unmarshalled VK response: %v", response)
	if response.Error != nil {
		err = response.Error
		log.Debugf(c, "VK API returned error - pass it upstream: %v", err)
	}
	return response.Response, err
}
