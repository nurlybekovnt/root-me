package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

const (
	// challengeURL is the URL of the challenge page
	challengeURL = "http://challenge01.root-me.org/programmation/ch1/"
	// submitURL is the URL to submit the solution
	submitURL = "http://challenge01.root-me.org/programmation/ch1/ep1_v.php?result="
)

func main() {
	challenge, err := Fetch()
	if err != nil {
		log.Panicf("failed to fetch challenge: %v", err)
	}

	solution := Solve(challenge)

	result, err := Submit(solution)
	if err != nil {
		log.Panicf("failed to submit solution: %v", err)
	}

	println(result)
}

// Un+1 = [ A + Un ] sign [ n * B ]
type Challenge struct {
	ZeroElement int64
	A, B        int64
	Sign        rune
	N           int

	cookies []*http.Cookie
}

var (
	regexpA = regexp.MustCompile(`\[\s+-?\d+\s+\+`)
	regexpB = regexp.MustCompile(`\*\s+-?\d+\s+\]`)
	regexpS = regexp.MustCompile(`\]\s+[+-]\s+\[`)
	regexpZ = regexp.MustCompile(`=\s+-?\d+`)
	regexpN = regexp.MustCompile(`>-?\d+<`)
)

func Fetch() (challenge Challenge, err error) {
	resp, err := http.Get(challengeURL)
	if err != nil {
		return Challenge{}, err
	}
	defer resp.Body.Close()

	lines, err := ReadLines(bufio.NewReader(resp.Body))
	if err != nil {
		return Challenge{}, err
	}

	if len(lines) != 3 {
		return Challenge{}, fmt.Errorf("invalid number of lines: %d", len(lines))
	}

	a := regexpA.FindString(lines[0])
	if a == "" {
		return Challenge{}, fmt.Errorf("failed to find A")
	}
	challenge.A, err = strconv.ParseInt(strings.TrimSpace(a[1:len(a)-1]), 10, 64)
	if err != nil {
		return Challenge{}, err
	}

	b := regexpB.FindString(lines[0])
	if b == "" {
		return Challenge{}, fmt.Errorf("failed to find B")
	}
	challenge.B, err = strconv.ParseInt(strings.TrimSpace(b[1:len(b)-1]), 10, 64)
	if err != nil {
		return Challenge{}, err
	}

	s := regexpS.FindString(lines[0])
	if s == "" {
		return Challenge{}, fmt.Errorf("failed to find Sign")
	}
	challenge.Sign = rune(strings.TrimSpace(s[1 : len(s)-1])[0])

	z := regexpZ.FindString(lines[1])
	if z == "" {
		return Challenge{}, fmt.Errorf("failed to find Zero Element")
	}
	challenge.ZeroElement, err = strconv.ParseInt(strings.TrimSpace(z[1:]), 10, 64)
	if err != nil {
		return Challenge{}, err
	}

	n := regexpN.FindString(lines[2])
	if n == "" {
		return Challenge{}, fmt.Errorf("failed to find N")
	}
	challenge.N, err = strconv.Atoi(strings.TrimSpace(n[1 : len(n)-1]))
	if err != nil {
		return Challenge{}, err
	}

	challenge.cookies = resp.Cookies()
	return
}

func ReadLines(r *bufio.Reader) ([]string, error) {
	var lines []string

	for {
		line, _, err := r.ReadLine()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		lines = append(lines, string(line))
	}

	return lines, nil
}

type Solution struct {
	Un int64

	cookies []*http.Cookie
}

func Solve(challenge Challenge) (solution Solution) {
	u := challenge.ZeroElement

	for n := 0; n < challenge.N; n++ {
		switch challenge.Sign {
		case '+':
			u = (challenge.A + u) + (int64(n) * challenge.B)
		case '-':
			u = (challenge.A + u) - (int64(n) * challenge.B)
		default:
			panic("invalid sign")
		}
	}

	return Solution{Un: u, cookies: challenge.cookies}
}

func Submit(solution Solution) (result string, err error) {
	url := submitURL + strconv.FormatInt(solution.Un, 10)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}

	for _, cookie := range solution.cookies {
		req.AddCookie(cookie)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}
