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

	freeDoctors := 0
	busyDoctors := 0
	patients := 0
	treated := 0

	// init un peu shlag
	switch *siteID {
	case 3:
		freeDoctors = 0
		patients = 5
	case 1:
		freeDoctors = 0
		patients = 1
	case 2:
		freeDoctors = 0
		patients = 2
	default:
		freeDoctors = 0
		patients = 0
	}

	scanner := bufio.NewScanner(os.Stdin)
	writer := bufio.NewWriter(os.Stdout)

	syncChan := make(chan struct{}, 1)
	syncChan <- struct{}{}

	// envoi de l'état initial au contrôleur
	fmt.Fprintf(writer, "UPDATE %d %d\n", *siteID, freeDoctors)
	writer.Flush()
	fmt.Fprintf(os.Stderr, "[APP %d] Médecins libres: %d | Malades: %d\n", *siteID, freeDoctors, patients)

	// demande initiale si besoin
	fmt.Fprintf(os.Stderr, "%d > %d\n", patients, freeDoctors)
	if patients > freeDoctors {
		time.Sleep(1 * time.Second)
		<-syncChan
		fmt.Fprintf(writer, "ASK %d %d\n", *siteID, 1)
		writer.Flush()
		fmt.Fprintf(os.Stderr, "[APP %d] ❓ Demande d'1 médecin envoyée\n", *siteID)
		syncChan <- struct{}{}
	}

	const red = "\033[31m"
	const reset = "\033[0m"

	// affichage périodique de l'état actuelle
	go func() {
		for {
			time.Sleep(10 * time.Second)
			<-syncChan
			fmt.Fprintf(os.Stderr, red+"[APP %d] État — médecins Libres: %d | médecins Occupés: %d | Malades: %d | Soignés: %d\n"+reset,
				*siteID, freeDoctors, busyDoctors, patients, treated)
			syncChan <- struct{}{}
		}
	}()

	// traitement des patients
	go func() {
		for {
			time.Sleep(1 * time.Second)
			<-syncChan

			if freeDoctors > 0 && patients > 0 {
				// Un médecin devient occupé
				freeDoctors--
				busyDoctors++
				fmt.Fprintf(os.Stderr, "[APP %d] 🏥 Traitement en cours...\n", *siteID)

				// Mise à jour immédiate du contrôleur (un médecin en moins dispo)
				fmt.Fprintf(writer, "UPDATE %d %d\n", *siteID, freeDoctors)
				writer.Flush()

				syncChan <- struct{}{}

				time.Sleep(5 * time.Second)

				<-syncChan
				busyDoctors--
				freeDoctors++
				patients--
				treated++

				fmt.Fprintf(os.Stderr, "[APP %d] Patient soigné. Reste %d malades |️ libres: %d\n",
					*siteID, patients, freeDoctors)

				// Mise à jour du contrôleur (le médecin est à nouveau dispo)
				fmt.Fprintf(writer, "UPDATE %d %d\n", *siteID, freeDoctors)
				writer.Flush()
			}

			syncChan <- struct{}{}
		}
	}()

	// réception des messages
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for scanner.Scan() {
			<-syncChan

			line := scanner.Text()
			fmt.Fprintf(os.Stderr, "[APP %d] Message contrôleur : %s\n", *siteID, line)

			tokens := strings.Fields(line)
			if len(tokens) > 0 && tokens[0] == "GIVE" && len(tokens) == 4 {
				dst, _ := strconv.Atoi(tokens[2])
				n, _ := strconv.Atoi(tokens[3])

				if dst == *siteID {
					freeDoctors += n
					fmt.Fprintf(os.Stderr, "[APP %d] Reçu %d médecin(s). Total libres: %d\n", *siteID, n, freeDoctors)

					// Mise à jour
					fmt.Fprintf(writer, "UPDATE %d %d\n", *siteID, freeDoctors)
					writer.Flush()
				}
			}

			syncChan <- struct{}{}
		}
	}()

	wg.Wait()
}
