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

	state := SiteState{
		ID:      *siteID,
		Doctors: 1,
		Known:   make(map[int]int),
	}

	scanner := bufio.NewScanner(os.Stdin)
	writer := bufio.NewWriter(os.Stdout)

	syncChan := make(chan struct{}, 1)
	syncChan <- struct{}{}

	fmt.Fprintf(os.Stderr, "[CTL %d] Prêt avec %d médecin(s)\n", state.ID, state.Doctors)

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		for scanner.Scan() {
			time.Sleep(2 * time.Second)
			<-syncChan // verrou : on bloque tout

			line := scanner.Text()
			fmt.Fprintf(os.Stderr, "[CTL %d] Reçu: %s\n", state.ID, line)

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
				// n, _ := strconv.Atoi(tokens[2]) // <- optionnel, ici on envoie que 1 médecin max

				fmt.Fprintf(os.Stderr, "[CTL %d] Nombre de médecin : %d\n", state.ID, state.Doctors)
				if state.Doctors > 0 {

					// On répond à la demande : envoie d’un médecin
					msg := fmt.Sprintf("GIVE %d %d %d", state.ID, fromID, 1)
					fmt.Fprintln(writer, msg)
					writer.Flush()
					fmt.Fprintf(os.Stderr, "[CTL %d] Envoi de 1 médecin à %d\n", state.ID, fromID)
				} else {
					// Pas assez de médecins → relayer
					fmt.Println(line)
					fmt.Fprintf(os.Stderr, "[CTL %d] Relais de la demande de %d\n", state.ID, fromID)
				}

			case "GIVE":
				if len(tokens) < 4 {
					syncChan <- struct{}{}
					continue
				}

				src, _ := strconv.Atoi(tokens[1])
				dst, _ := strconv.Atoi(tokens[2])
				n, _ := strconv.Atoi(tokens[3])

				if dst == state.ID {
					state.Doctors += n
					fmt.Fprintf(os.Stderr, "[CTL %d] Reçu %d médecin(s) de %d\n", state.ID, n, src)
				} else {
					// Relais du message
					fmt.Println(line)
					fmt.Fprintf(os.Stderr, "[CTL %d] Relais de GIVE vers %d\n", state.ID, dst)
				}

			case "UPDATE":
				if len(tokens) < 3 {
					syncChan <- struct{}{}
					continue
				}
				id, _ := strconv.Atoi(tokens[1])
				n, _ := strconv.Atoi(tokens[2])
				state.Known[id] = n
				if id == state.ID {
					state.Doctors = n
				}
			}

			syncChan <- struct{}{} // relâcher le verrou
		}
	}()

	wg.Wait()
}
