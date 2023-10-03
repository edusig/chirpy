package main

import "strings"

func getProfaneWords() []string {
	return []string{"kerfuffle", "sharbert", "fornax"}
}

func cleanBody(body string) string {
	profaneWords := getProfaneWords()
	words := strings.Split(body, " ")
	cleanedBody := make([]string, 0)
	for _, word := range words {
		if contains(profaneWords, strings.ToLower(word)) {
			cleanedBody = append(cleanedBody, "****")
		} else {
			cleanedBody = append(cleanedBody, word)
		}
	}
	return strings.Join(cleanedBody, " ")
}

func contains(arr []string, search string) bool {
	for _, v := range arr {
		if v == search {
			return true
		}
	}
	return false
}
