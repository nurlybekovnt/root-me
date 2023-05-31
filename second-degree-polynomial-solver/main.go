package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"log"
	"math"
	"net"
	"regexp"
	"strconv"
	"strings"

	"github.com/nurlybekovnt/root-me-programmation/utils"
)

var (
	challengeSrvAddr = flag.String("addr", "challenge01.root-me.org:52018", "challenge server address")
	totalChallenges  = flag.Int("total", 25, "total number of challenges to solve")
)

func main() {
	flag.Parse()

	conn, err := net.Dial("tcp", *challengeSrvAddr)
	if err != nil {
		log.Panicf("failed to connect to the server: %v", err)
	}
	defer conn.Close()

	client := &Client{
		conn:   conn,
		buffer: make([]byte, 1024),
	}

	for i := 0; i < *totalChallenges; i++ {
		challenge, err := client.FetchChallenge()
		if err != nil {
			log.Panicf("failed to fetch challenge: %v", err)
		}

		solution := Solve(challenge)
		log.Printf("%d/%d challenge: %+v, solution: %+v",
			i+1,
			*totalChallenges,
			challenge,
			solution,
		)

		if err := client.SubmitSolution(solution); err != nil {
			log.Panicf("failed to submit solution: %v", err)
		}
	}

	data, err := client.Read()
	if err != nil {
		log.Panicf("failed to read from the server: %v", err)
	}
	log.Printf("server response: %s", data)
}

type Client struct {
	conn   net.Conn
	buffer []byte
}

type Challenge struct {
	A, B, C int
}

type Solution struct {
	Roots []float64
}

var (
	regexpNumber = regexp.MustCompile(`[+-]?\s?\d+`)
)

func (c *Client) Read() ([]byte, error) {
	n, err := c.conn.Read(c.buffer)
	if err != nil {
		return nil, fmt.Errorf("failed to read from the server: %w", err)
	}

	return c.buffer[:n], nil
}

func (c *Client) FetchChallenge() (ch Challenge, err error) {
	data, err := c.Read()
	if err != nil {
		return Challenge{}, fmt.Errorf("failed to read from the server: %w", err)
	}

	lines, err := utils.ReadLines(bufio.NewReader(bytes.NewReader(data)))
	if err != nil {
		return Challenge{}, fmt.Errorf("failed to read lines: %w", err)
	}

	if len(lines) < 2 {
		return Challenge{}, fmt.Errorf("invalid number of lines: %d", len(lines))
	}

	_, equation, ok := strings.Cut(lines[len(lines)-2], ": ")
	if !ok {
		return Challenge{}, fmt.Errorf("failed to parse equation: %w", err)
	}

	tokens := regexpNumber.FindAllString(equation, 4)
	if len(tokens) != 4 {
		return Challenge{}, fmt.Errorf("invalid number of integers: %d", len(tokens))
	}

	numbers := make([]int, len(tokens))
	for i := range tokens {
		numbers[i], err = strconv.Atoi(strings.ReplaceAll(tokens[i], " ", ""))
		if err != nil {
			return Challenge{}, fmt.Errorf("failed to parse integer: %w", err)
		}
	}

	return Challenge{
		A: numbers[0],
		B: numbers[1],
		C: numbers[2] - numbers[3],
	}, nil
}

func format(f float64) string {
	s := fmt.Sprintf("%.3f", f)
	s = strings.TrimRight(s, "0")
	s = strings.TrimRight(s, ".")
	return s
}

func (c *Client) SubmitSolution(s Solution) error {
	var response string

	switch len(s.Roots) {
	case 0:
		response = "Not possible"
	case 1:
		response = fmt.Sprintf("x: %s", format(s.Roots[0]))
	case 2:
		response = fmt.Sprintf("x1: %s ; x2: %s", format(s.Roots[0]), format(s.Roots[1]))
	default:
		return fmt.Errorf("invalid number of roots: %d", len(s.Roots))
	}

	_, err := fmt.Fprintln(c.conn, response)
	return err
}

func Solve(ch Challenge) (s Solution) {
	discriminant := float32(ch.B*ch.B - 4*ch.A*ch.C)

	switch {
	case discriminant > 0:
		sqrtD := math.Sqrt(float64(discriminant))
		return Solution{
			Roots: []float64{
				(-float64(ch.B) + sqrtD) / (2 * float64(ch.A)),
				(-float64(ch.B) - sqrtD) / (2 * float64(ch.A)),
			},
		}
	case discriminant == 0:
		return Solution{
			Roots: []float64{
				float64(-ch.B) / (2 * float64(ch.A)),
			},
		}
	default:
		return
	}
}
