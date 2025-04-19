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

func main() {
	siteID := flag.Int("n", 0, "ID du site")
	flag.Parse()
	doctors := 0
	patients := 0

	switch *siteID {
	case 3:
		doctors = 0
		patients = 5
	case 1:
		doctors = 1
		patients = 1
	case 2:
		doctors = 3
		patients = 2
	default:
		doctors = 1
		patients = 1
	}

	scanner := bufio.NewScanner(os.Stdin)
	writer := bufio.NewWriter(os.Stdout)

	syncChan := make(chan struct{}, 1)
	syncChan <- struct{}{}

	// État initial
	fmt.Fprintf(writer, "UPDATE %d %d\n", *siteID, doctors)
	writer.Flush()
	fmt.Fprintf(os.Stderr, "[APP %d] Médecins: %d, Malades: %d\n", *siteID, doctors, patients)

	// Envoi d'une demande initiale si besoin
	if patients > doctors {
		time.Sleep(1 * time.Second)
		<-syncChan
		fmt.Fprintf(writer, "ASK %d %d\n", *siteID, 1) // ASK <fromID> <nombre>
		writer.Flush()
		fmt.Fprintf(os.Stderr, "[APP %d] Envoi d'une demande de 1 médecin\n", *siteID)
		syncChan <- struct{}{}
	}

	const red = "\033[31m"
	const reset = "\033[0m"

	// Affichage périodique en rouge toutes les 10 secondes
	go func() {
		for {
			time.Sleep(10 * time.Second)
			<-syncChan
			fmt.Fprintf(os.Stderr, red+"[APP %d] 🔁 État périodique — Médecins: %d, Malades: %d\n"+reset, *siteID, doctors, patients)
			syncChan <- struct{}{}
		}
	}()

	// Traitement des patients
	go func() {
		for {
			time.Sleep(1 * time.Second)
			<-syncChan

			if doctors > 0 && patients > 0 {
				doctors--
				fmt.Fprintf(os.Stderr, "[APP %d] 🏥 Traitement en cours...\n", *siteID)
				syncChan <- struct{}{}

				time.Sleep(5 * time.Second)

				<-syncChan
				doctors++
				patients--
				fmt.Fprintf(os.Stderr, "[APP %d] ✅ Patient soigné. Reste %d malades, %d médecins\n", *siteID, patients, doctors)

				// Mise à jour du contrôleur
				fmt.Fprintf(writer, "UPDATE %d %d\n", *siteID, doctors)
				writer.Flush()
			}

			syncChan <- struct{}{}
		}
	}()

	// Réception des messages du contrôleur
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for scanner.Scan() {
			<-syncChan

			line := scanner.Text()
			fmt.Fprintf(os.Stderr, "[APP %d] Contrôleur dit : %s\n", *siteID, line)

			tokens := strings.Fields(line)
			if len(tokens) > 0 && tokens[0] == "GIVE" && len(tokens) == 4 {
				dst, _ := strconv.Atoi(tokens[2])
				n, _ := strconv.Atoi(tokens[3])

				if dst == *siteID {
					doctors += n
					fmt.Fprintf(os.Stderr, "[APP %d] 🚑 Reçu %d médecin(s) ! Total: %d\n", *siteID, n, doctors)

					// Mise à jour du contrôleur
					fmt.Fprintf(writer, "UPDATE %d %d\n", *siteID, doctors)
					writer.Flush()
				}
			}

			syncChan <- struct{}{}
		}
	}()

	wg.Wait()
}
