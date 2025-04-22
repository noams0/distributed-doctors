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

func updateCtrl(writer *bufio.Writer, siteID int, freeDoctors int) {
	fmt.Fprintf(writer, "UPDATE %d %d\n", siteID, freeDoctors)
	writer.Flush()
}

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
		freeDoctors = 2
		patients = 1
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
	updateCtrl(writer, *siteID, freeDoctors)

	fmt.Fprintf(os.Stderr, "[APP %d] Médecins libres: %d | Malades: %d\n", *siteID, freeDoctors, patients)

	if patients > freeDoctors {
		time.Sleep(1 * time.Second)
		<-syncChan
		fmt.Fprintf(writer, "ASK %d %d\n", *siteID, 1)
		writer.Flush()
		fmt.Fprintf(os.Stderr, "[APP %d] ❓ Demande d'1 médecin envoyée\n", *siteID)
		syncChan <- struct{}{}
	}

	const Reset = "\033[0m"
	const Red = "\033[31m"
	const Green = "\033[32m"
	const Yellow = "\033[33m"
	const Blue = "\033[34m"
	const Purple = "\033[35m"
	const Cyan = "\033[36m"
	const Gray = "\033[37m"
	const White = "\033[97m"
	// affichage périodique de l'état actuelle
	go func() {
		for {
			time.Sleep(10 * time.Second)
			<-syncChan
			fmt.Fprintf(os.Stderr, Blue+"[APP %d] État — médecins Libres: %d | médecins Occupés: %d | Malades: %d | Soignés: %d\n"+Reset,
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
				fmt.Fprintf(os.Stderr, Red+"[APP %d] 1 docteur libre en moins, Traitement en cours...\n"+Reset, *siteID)

				// Mise à jour immédiate du contrôleur (un médecin en moins dispo)
				updateCtrl(writer, *siteID, freeDoctors)

				syncChan <- struct{}{}

				time.Sleep(5 * time.Second)

				<-syncChan
				busyDoctors--
				freeDoctors++
				patients--
				treated++

				fmt.Fprintf(os.Stderr, Green+"[APP %d] Médecin de nouveau dispo. Patient soigné. Reste %d malades |️ libres: %d\n"+Reset,
					*siteID, patients, freeDoctors)

				// Mise à jour du contrôleur (le médecin est à nouveau dispo)
				updateCtrl(writer, *siteID, freeDoctors)

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
			if len(tokens) > 0 && tokens[0] == "ASK" {
				src, _ := strconv.Atoi(tokens[1])
				n, _ := strconv.Atoi(tokens[2])
				if freeDoctors > 0 {
					// soit on répond à la demande : envoie d’un médecin
					msg := fmt.Sprintf("GIVE %d %d %d", *siteID, src, n)
					fmt.Fprintln(writer, msg)
					writer.Flush()
					freeDoctors -= n
					fmt.Fprintf(os.Stderr, Red+"[APP %d] Envoie de %d médecin(s). Total libres: %d\n"+Reset, *siteID, n, freeDoctors)
					updateCtrl(writer, *siteID, freeDoctors)
				}
			}
			if len(tokens) > 0 && tokens[0] == "GIVE" && len(tokens) == 4 {
				//src, _ := strconv.Atoi(tokens[1])
				dst, _ := strconv.Atoi(tokens[2])
				n, _ := strconv.Atoi(tokens[3])

				if dst == *siteID {
					freeDoctors += n
					fmt.Fprintf(os.Stderr, Green+"[APP %d] Reçu %d médecin(s). Total libres: %d\n"+Reset, *siteID, n, freeDoctors)
					updateCtrl(writer, *siteID, freeDoctors)
				}
			}

			syncChan <- struct{}{}
		}
	}()

	wg.Wait()
}
