package services

import (
	"fmt"
	"regexp"
	"strings"

	"transcription-goserver/internal/logging"
)

type QualityChecker struct{}

var nonWordRegex = regexp.MustCompile(`[.,!?"']`)

func (qc *QualityChecker) CheckAudioQuality(transcript, callID string) (bool, string) {
	logStreamer := logging.GetLogStreamer()

	if transcript == "" || len(strings.TrimSpace(transcript)) == 0 {
		return false, "empty_transcript"
	}

	text := strings.ToLower(strings.TrimSpace(transcript))
	words := strings.Fields(text)
	totalWords := len(words)

	if totalWords < 5 {
		logStreamer.Debug("QualityChecker", fmt.Sprintf("Call %s: Short transcript (%d words)", callID, totalWords))
		return true, "short_allowed"
	}

	cleanedWords := make([]string, 0)
	for _, w := range words {
		cw := nonWordRegex.ReplaceAllString(strings.ToLower(w), "")
		if cw != "" {
			cleanedWords = append(cleanedWords, cw)
		}
	}

	if len(cleanedWords) == 0 {
		return false, "no_valid_words"
	}

	wordCounts := make(map[string]int)
	for _, w := range cleanedWords {
		wordCounts[w]++
	}

	var mostCommonWord string
	var mostCommonCount int
	for w, c := range wordCounts {
		if c > mostCommonCount {
			mostCommonCount = c
			mostCommonWord = w
		}
	}

	frequencyRatio := float64(mostCommonCount) / float64(len(cleanedWords))
	logStreamer.Debug("QualityChecker",
		fmt.Sprintf("Call %s: Most common '%s' = %.1f%%", callID, mostCommonWord, frequencyRatio*100))

	if frequencyRatio >= 0.80 {
		return false, fmt.Sprintf("repetitive_%s_%.0fpct", mostCommonWord, frequencyRatio*100)
	}

	if len(wordCounts) >= 2 {
		counts := make([]int, 0, len(wordCounts))
		for _, c := range wordCounts {
			counts = append(counts, c)
		}
		for i := 0; i < len(counts); i++ {
			for j := i + 1; j < len(counts); j++ {
				if counts[j] > counts[i] {
					counts[i], counts[j] = counts[j], counts[i]
				}
			}
		}

		topTwoCount := counts[0]
		if len(counts) > 1 {
			topTwoCount += counts[1]
		}
		topTwoRatio := float64(topTwoCount) / float64(len(cleanedWords))
		uniqueWords := len(wordCounts)

		if topTwoRatio >= 0.90 && uniqueWords <= 4 {
			return false, fmt.Sprintf("low_variety_%d_unique", uniqueWords)
		}
	}

	consecutiveCount := 1
	maxConsecutive := 1
	for i := 1; i < len(cleanedWords); i++ {
		if cleanedWords[i] == cleanedWords[i-1] && len(cleanedWords[i]) > 0 {
			consecutiveCount++
			if consecutiveCount > maxConsecutive {
				maxConsecutive = consecutiveCount
			}
		} else {
			consecutiveCount = 1
		}
	}

	if maxConsecutive >= 10 {
		return false, fmt.Sprintf("consecutive_repeat_%d", maxConsecutive)
	}

	if totalWords >= 20 {
		uniqueRatio := float64(len(wordCounts)) / float64(len(cleanedWords))
		if uniqueRatio < 0.1 {
			return false, fmt.Sprintf("low_unique_%.0fpct", uniqueRatio*100)
		}
	}

	return true, "clear_speech"
}

func (qc *QualityChecker) CleanTranscript(transcript string) string {
	if transcript == "" {
		return transcript
	}

	words := strings.Fields(transcript)
	if len(words) < 2 {
		return transcript
	}

	cleaned := make([]string, 0, len(words))
	cleaned = append(cleaned, words[0])
	consecutive := 1

	for i := 1; i < len(words); i++ {
		currentClean := nonWordRegex.ReplaceAllString(strings.ToLower(words[i]), "")
		prevClean := nonWordRegex.ReplaceAllString(strings.ToLower(words[i-1]), "")

		if currentClean == prevClean {
			consecutive++
			if consecutive <= 4 {
				cleaned = append(cleaned, words[i])
			}
		} else {
			consecutive = 1
			cleaned = append(cleaned, words[i])
		}
	}

	return strings.Join(cleaned, " ")
}
