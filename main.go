package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
)

func sendToChannel(w http.ResponseWriter, s string) {
	var answer slackAnswer
	answer.ResponseType="in_channel"
	answer.Text=s
	json.NewEncoder(w).Encode(answer)
}

func rules(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "*Rules* Everyone starts at 21 pts. You gain 2 pts per perfect week. You loose pts based on how many minutes you are late to a meeting. You can buy back points by proposing stuff to the team and them telling you how much that is worth." +
		"\n*Exceptions* You won't be counted as being late if you are actively fixing prod. You won't be counted as being late if you can't attend and post your status update in time.")
}

func leaderboard(w http.ResponseWriter, r *http.Request) {
	//TODO Would be cooler to print by points instead of by name
	b := new(bytes.Buffer)
	names := make([]string, 0)
	for name, _ := range members {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		fmt.Fprintf(b, "%s has %s points.\n", name, members[name])
	}

	w.WriteHeader(http.StatusOK)
	sendToChannel(w, b.String())
}

func resetAll(w http.ResponseWriter, r *http.Request) {
	for name := range members {
		members[name] = 21
	}

	leaderboard(w, r)
}

func handleMember(w http.ResponseWriter, r *http.Request) {
	command := mux.Vars(r)["command"]
	name := mux.Vars(r)["text"]
	if strings.Contains(command, "add") {
		addMember(w, name)
	} else if strings.Contains(command, "rm") {
		removeMember(w, name)
	}
}

func addMember(w http.ResponseWriter, name string) {
	isAlpha := regexp.MustCompile(`^[A-Za-z]+$`).MatchString
	if !isAlpha(name) {
		fmt.Fprintf(w, "Invalid name [%s] contains non alphabetic characters.", name)
		return
	}

	if _, ok := members[name]; ok {
		w.WriteHeader(http.StatusConflict)
		fmt.Fprintf(w, "Member %s is already part of timekeeper bot.", name)
	} else {
		members[name]=21

		w.WriteHeader(http.StatusCreated)
		sendToChannel(w, "Member " + name + " has been added to timekeeper bot. Good luck!")
	}
}

func removeMember(w http.ResponseWriter, name string) {
	if _, ok := members[name]; ok {
		delete(members, name)

		w.WriteHeader(http.StatusNoContent)
		sendToChannel(w, "Member " + name + " has been removed from timekeeper bot.")
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}

func assignPoints(w http.ResponseWriter, r *http.Request) {
	text := mux.Vars(r)["text"]
	name := strings.Split(text, " ")[0]
	points, err := strconv.Atoi(strings.Split(text, " ")[1])

	if err != nil {
		fmt.Fprintf(w, "Please make sure the points are an actual integer.")
		return
	}

	if _, ok := members[name]; ok {
		members[name] = members[name] + points
		if members[name] > 21 {
			members[name] = 21
		}

		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Member %s now at %s points.", name, strconv.FormatInt(int64(members[name]), 10))
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}

func removePoints(w http.ResponseWriter, r *http.Request) {
	text := mux.Vars(r)["text"]
	name := strings.Split(text, " ")[0]
	minutes, err := strconv.Atoi(strings.Split(text, " ")[1])

	if err != nil {
		fmt.Fprintf(w, "Please make sure the minutes are an actual integer.")
		return
	}

	if _, ok := members[name]; ok {
		members[name] = members[name] - getPoints(minutes)

		w.WriteHeader(http.StatusOK)
		if members[name] < 0 {
			members[name] = 0
			sendToChannel(w, "Uh-oh! Member " + name + " no longer has points :scream:")
		} else {
			fmt.Fprintf(w, "Member %s now at %s points.", name, strconv.FormatInt(int64(members[name]), 10))
		}
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}

func getPoints(timeLate int) (points int) {
	if timeLate < 1 { return 0 }
	if timeLate < 2 { return 2 }
	if timeLate < 3 { return 3 }
	if timeLate < 5 { return 4 }
	if timeLate < 8 { return 5 }
	if timeLate < 13 { return 6 }
	if timeLate < 21 { return 7 }
	return 8
}

func perfectWeek(w http.ResponseWriter, r *http.Request) {
	names := strings.Split(mux.Vars(r)["text"], ",")

	for _, name := range names {
		members[name] = members[name] + 2
	}

	leaderboard(w, r)
}

func main() {
	router := mux.NewRouter().StrictSlash(true)
	// See section 2 for detail on slack payload: https://api.slack.com/interactivity/slash-commands
	//TODO Add request validation with https://api.slack.com/docs/verifying-requests-from-slack
	router.HandleFunc("/rules", rules).Methods("POST")
	router.HandleFunc("/leaderboard", leaderboard).Methods("POST")
	router.HandleFunc("/leaderboard/reset", resetAll).Methods("POST")
	router.HandleFunc("/member/{name}", handleMember).Methods("POST")
	router.HandleFunc("/member/{name}/points/{points}", assignPoints).Methods("POST")
	router.HandleFunc("/member/{name}/late/{minutes}", removePoints).Methods("POST")
	router.HandleFunc("/perfect_week", perfectWeek).Methods("POST")
	log.Fatal(http.ListenAndServe(":8080", router))
}

type slackAnswer struct {
	ResponseType	string `json:"response_type"`
	Text			string `json:"text"`
}

//TODO Implement proper persistency
var members = map[string] int {
}