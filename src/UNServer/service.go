package main

import (
	l4g "classified-lib/golang-lib/log"
	ctx "context"
	"daomanage"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"service/auth"
	"strconv"
)

const (
	InfoLoginFailed = "Incorrect username or password"
	InfoLoginSucc   = "Let's GO"

	InfoHack = "What the fk r u looking for?"

	PlbPerPage = 50
)

type ErrorInfo struct {
	Status string
	Info   string
}

func isAuthorized(query url.Values) bool {
	var ipaddr string

	// Authorize this query
	if len(query["ipaddr"]) > 0 {
		ipaddr = query["ipaddr"][0]
	} else {
		return false
	}
	authCode, authErr := ctx.SvrCtx.AuthMan.DoAuth(ipaddr)
	if authErr != nil {
		e := fmt.Errorf("Faild to auth this request for some internal error: %s", authErr)
		l4g.WarnLogger.Warn("%s", e)
		return false
	}
	if authCode != auth.AuthCode_OK {
		e := fmt.Errorf("Failed to auth this request, authCode: %d", authCode)
		l4g.WarnLogger.Warn("%s", e)
		return false
	}
	return true
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	var (
		query     url.Values
		err       error
		userInfo  daomanage.User
		errorInfo ErrorInfo
	)
	query, err = url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		log.Println("Parse Error", err)
		return
	}

	// Get user info
	if s, ok := query["username"]; ok {
		userInfo, err = daomanage.GetUserInfo("username", s[0])
		if p, ok := query["password"]; ok {
			if p[0] == userInfo.Password {
				errorInfo.Status = "OK"
				errorInfo.Info = InfoLoginSucc
				data, _ := json.Marshal(errorInfo)
				fmt.Fprintf(w, string(data))
				return
			}
		}
		errorInfo.Status = "Error"
		errorInfo.Info = InfoLoginFailed
		data, _ := json.Marshal(errorInfo)
		fmt.Fprintf(w, string(data))
		return
	}
}

func problemInfoHandler(w http.ResponseWriter, r *http.Request) {
	var (
		query       url.Values
		err         error
		problemInfo daomanage.Problem
	)
	query, err = url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		log.Println("Parse Error", err)
		return
	}

	// Get problem info
	if s, ok := query["pid"]; ok {
		pid, err := strconv.ParseInt(s[0], 10, 64)
		if err != nil {
			return
		}
		problemInfo, err = daomanage.GetProblemInfo(pid)
		if err == nil {
			data, _ := json.Marshal(problemInfo)
			fmt.Fprintf(w, string(data))
			return
		}
	}
	fmt.Fprintf(w, InfoHack)
}

func problemsHandler(w http.ResponseWriter, r *http.Request) {
	var (
		query url.Values
		err   error
	)
	query, err = url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		log.Println("Parse Error", err)
		return
	}

	var page int64
	if s, ok := query["page"]; ok {
		if s[0] == "" {
			page = 1
		} else {
			page, err = strconv.ParseInt(s[0], 10, 64)
			if err != nil {
				log.Println("re")
				return
			}
		}
	}
	startIndex := PlbPerPage * (page - 1)
	endIndex := startIndex + PlbPerPage

	// Get Problems
	problemInfoList, err := daomanage.GetProblemInRange(startIndex, endIndex)
	if err == nil {
		data, _ := json.Marshal(problemInfoList)
		fmt.Fprintf(w, string(data))
		return
	}
}

func contestsHandler(w http.ResponseWriter, r *http.Request) {
	var (
		// query       url.Values
		err         error
		contestInfo daomanage.Contest
	)
	// query, err = url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		log.Println("Parse Error", err)
		return
	}

	// Get contests info
	contestInfo, err = daomanage.GetContest(1)
	if err == nil {
		data, _ := json.Marshal(contestInfo)
		fmt.Fprintf(w, string(data))
		return
	}
	fmt.Fprintf(w, InfoHack)
}

func contestInfoHandler(w http.ResponseWriter, r *http.Request) {
	var (
		query       url.Values
		err         error
		contestInfo []daomanage.ContestProblem
	)

	type ContestProblem struct {
		CID         int64  `bson:"cid" json:"cid"`
		PID         int64  `bson:"pid" json:"pid"`
		ProblemName string `bson:"problemname" json:"problemname"`
		Solved      int64  `bson:"solved" json:"solved"`
		Score       int64  `bson:"score" json:"score"`
	}
	var contestProblemList []ContestProblem

	query, err = url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		log.Println("Parse Error", err)
		return
	}

	// Get contest info
	if s, ok := query["cid"]; ok {
		cid, err := strconv.ParseInt(s[0], 10, 64)
		if err != nil {
			return
		}
		contestInfo, err = daomanage.GetContestProblems(cid)
		if err == nil {
			// data, _ := json.Marshal(contestInfo)
			for _, problem := range contestInfo {
				problemInfo, err := daomanage.GetProblemInfo(problem.PID)
				if err != nil {
					continue
				}
				singleProblemInfo := ContestProblem{problem.CID, problem.PID, problemInfo.Title, problem.Solved, problem.Score}
				contestProblemList = append(contestProblemList, singleProblemInfo)
			}
			data, _ := json.Marshal(contestProblemList)
			fmt.Printf("%s\n", data)
			fmt.Fprintf(w, string(data))
			return
		}
	}
	fmt.Fprintf(w, InfoHack)
}

func userHandler(w http.ResponseWriter, r *http.Request) {
	var (
		query    url.Values
		err      error
		userInfo daomanage.User
	)
	query, err = url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		log.Println("Parse Error", err)
		return
	}

	// ok := isAuthorized(query)
	// if ok == false {
	// 	fmt.Fprintf(w, InfoHack)
	// 	return
	// }

	// Get user info
	if s, ok := query["uid"]; ok {
		userInfo, err = daomanage.GetUserInfo("uid", s[0])
		if err == nil {
			data, _ := json.Marshal(userInfo)
			fmt.Fprintf(w, string(data))
			return
		}
	}
	fmt.Fprintf(w, InfoHack)
}
