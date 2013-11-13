package main

import (
	"time"
	"fmt"
	"strings"
	"encoding/json"
	"bytes"
	"log"
)

func init(){
	parser.AddCommand("log",
		"Manipulate time log",
		"The log command permits the listing and logging of time on specific stories.",
		&logCommand)

}

type LogCommand struct {
	MyLog bool `short:"m" long:"mine" description:"Show my log for current sprint"`
	Author string `short:"a" long:"author" description:"Show log for given author"`
	jc JiraClient
}

var logCommand LogCommand

type TimeLog struct {
	Key     string
	Seconds int
}

func (tl TimeLog) String() string {
	t, _ := time.ParseDuration(fmt.Sprintf("%ds", tl.Seconds))

	return fmt.Sprintf("%s : %s", tl.Key, fmt.Sprint(t))
}

type Period struct{
	Begin time.Time
	End time.Time
}

func (lc *LogCommand) GetTimeLog(targetAuthor string, period Period, issue *Issue){
		lastsundaybeforeperiod, lastsaturdaybeforeperiod := time.Date(2013, 11, 10, 0, 0, 0, 0, time.Local), time.Date(2013, 11, 16, 0, 0, 0, 0, time.Local)
	issuestring := ""
	if issue != nil {
		issuestring = fmt.Sprintf(" AND key = '%s'", issue.Key)
	}
	issues, _ := lc.jc.Search(&SearchOptions{JQL: fmt.Sprintf("timespent > 0 AND updated >= '%s' AND updated <= '%s' and project = '%s'%s", period.Begin.Format("2006-01-02"), period.End.Format("2006-01-02"), options.Project, issuestring)})
		logs_for_times := map[time.Time][]TimeLog{}
		for _, issue := range issues {
			url := fmt.Sprintf("https://%s:%s@jira.gammae.com/rest/api/2/issue/%s/worklog", options.User, options.Passwd, issue.Key)
			resp, _ := lc.jc.client.Get(url)
			worklog, _ := JsonToInterface(resp.Body)
			logs_json, _ := jsonWalker("worklogs", worklog)
			logs, ok := logs_json.([]interface{})
			if ok {
				for _, log := range logs {
					//We got good json and it's by our user
					authorjson, _ := jsonWalker("author/name", log)
					if author, ok := authorjson.(string); ok && (author == targetAuthor || targetAuthor == "") {
						dsjson, _ := jsonWalker("started", log)
						if date_string, ok := dsjson.(string); ok {
							//"2013-11-08T11:37:03.000-0500" <-- date format
							precise_time, _ := time.Parse("2006-01-02T15:04:05.000-0700", date_string)
							if precise_time.After(lastsundaybeforeperiod) && precise_time.Before(lastsaturdaybeforeperiod) {
								date := time.Date(precise_time.Year(), precise_time.Month(), precise_time.Day(), 0, 0, 0, 0, precise_time.Location())
								secondsjson, _ := jsonWalker("timeSpentSeconds", log)
								seconds := int(secondsjson.(float64))
								if _, ok := logs_for_times[date]; !ok {
									logs_for_times[date] = make([]TimeLog, 0)
								}
								logs_for_times[date] = append(logs_for_times[date], TimeLog{issue.Key, seconds})
							}
						}
					}
				}
			}
		}
		for t, l := range logs_for_times {
			fmt.Println(t)
			for _, singlelog := range l {
				fmt.Println(singlelog)
			}
		}

}

func (lc *LogCommand) Execute(args []string) error {
	jc := NewJiraClient(options)
	if lc.MyLog || len(args) < 2 {
		author := ""
		if lc.MyLog {
			author = options.User
		}
		if lc.Author != ""{
			author = lc.Author
		}
		var issue *Issue = nil
		if len(args) > 0 {
			issue = &Issue{Key:args[0]}
		}
		n := time.Now()
		beg := time.Date(n.Year(), n.Month(), n.Day() - int(n.Weekday()), 0, 0, 0, 0, n.Location())
		end := time.Date(n.Year(), n.Month(), n.Day() - int(n.Weekday()) + 6 , 0, 0, 0, 0, n.Location())
		lc.GetTimeLog(author, Period{beg, end}, issue)
	} else {
		key := args[0]
		time := strings.Join(args[1:], " ")

		postdata, _ := json.Marshal(map[string]string{"timeSpent": time})

		url := fmt.Sprintf("https://%s:%s@jira.gammae.com/rest/api/2/issue/%s/worklog", options.User, options.Passwd, key)
		resp, err := jc.client.Post(url, "application/json", bytes.NewBuffer(postdata))
		if err != nil {
			panic(err)
		}
		if resp.StatusCode == 201 {
			log.Println("Log successful")
		} else {
			log.Println("Log Failed!")
		}
	}
	return nil
}