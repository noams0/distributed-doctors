package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type SiteState struct {
	ID      int
	Doctors int
	Known   map[int]int
}

func main() {
	siteID := flag.Int("n", 0, "ID du site")
	flag.Parse()

	scanner := bufio.NewScanner(os.Stdin)
	writer := bufio.NewWriter(os.Stdout)

	syncChan := make(chan struct{}, 1)
	syncChan <- struct{}{}
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		for scanner.Scan() {
			time.Sleep(2 * time.Second)
			<-syncChan // verrou : on bloque tout

			line := scanner.Text()
			fmt.Fprintf(os.Stderr, "[CTL %d] Reçu: %s\n", *siteID, line)

			tokens := strings.Fields(line)
			if len(tokens) == 0 {
				syncChan <- struct{}{}
				continue
			}

			switch tokens[0] {
			case "ASK":
				if len(tokens) < 3 {
					syncChan <- struct{}{}
					continue
				}

				fromID, _ := strconv.Atoi(tokens[1])
				// n, _ := strconv.Atoi(tokens[2]) //si on voudra faire transier plusieurs médecins

				//DIFFERENCIER QUAND ON ENVIUE MESSAGE AU CTL VS APPLICATION
				fmt.Fprintf(os.Stderr, "[CTL %d] Relais de la demande de %d\n", *siteID, fromID)
				fmt.Println(line)
			case "GIVE":
				if len(tokens) < 4 {
					syncChan <- struct{}{}
					continue
				}

				src, _ := strconv.Atoi(tokens[1])
				dst, _ := strconv.Atoi(tokens[2])
				n, _ := strconv.Atoi(tokens[3])

				if dst == *siteID {
					//state.Doctors += n => pas à ctl de gérer ça
					fmt.Fprintf(os.Stderr, "[CTL %d] Reçu %d médecin(s) de %d\n", *siteID, n, src)
					msg := fmt.Sprintf("GIVE %d %d %d", src, dst, n)
					fmt.Fprintln(writer, msg)
					writer.Flush()

				} else {
					// Relais du message
					fmt.Println(line)
					fmt.Fprintf(os.Stderr, "[CTL %d] Relais de GIVE vers %d\n", *siteID, dst)
				}

			case "UPDATE":
				if len(tokens) < 3 {
					syncChan <- struct{}{}
					continue
				}
			}

			syncChan <- struct{}{} // release du verrou
		}
	}()

	wg.Wait()
}
