package main

import (
	"encoding/json"
	"fmt"
	"github.com/cloudfoundry/gosigar"
	"gopkg.in/alecthomas/kingpin.v1"
	"log"
	"math"
	"net"
	"os"
	"time"
)

var (
	interval   = kingpin.Flag("interval", "Monitoring Interval").Default("1s").Short('i').Duration()
	graphite   = kingpin.Flag("server", "Graphite server hostname:port").Short('s').OverrideDefaultFromEnvar("GRAPHITE_SERVER").String()
	showLoad   = kingpin.Flag("load", "Send Load data").Short('l').Bool()
	showCPU    = kingpin.Flag("cpu", "Send CPU data").Short('c').Bool()
	showMemory = kingpin.Flag("memory", "Send Memory data").Short('m').Bool()
	filesystem = kingpin.Flag("filesystem", "Send Filesystem data for specified path").Short('f').String()
)

type CPUSeries struct {
	Name    string   `json:"name"`
	Columns []string `json:"columns"`
	Points  [1][]int `json:"points"`
}

type MemSeries struct {
	Name    string   `json:"name"`
	Columns []string `json:"columns"`
	Points  [1][]int `json:"points"`
}

type DiskSeries struct {
	Name    string   `json:"name"`
	Columns []string `json:"columns"`
	Points  [1][]int `json:"points"`
}

type LoadSeries struct {
	Name    string       `json:"name"`
	Columns []string     `json:"columns"`
	Points  [1][]float64 `json:"points"`
}

var conn *net.UDPConn
var addr *net.UDPAddr

func connectToGraphite() {
	var err error

	addr, err = net.ResolveUDPAddr("udp", *graphite)

	if err != nil {
		panic(err)
	}

	conn, err = net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4zero, Port: 0})

	if err != nil {
		panic(err)
	}
}

func round(f float64) float64 {
	return math.Floor(f + .5)
}

func roundLoad(f float64) float64 {
	shift := math.Pow(10, float64(2))
	return round(f*shift) / shift
}

func writeData(jsonOutput []byte) (err error) {
	updates := []map[string]interface{}{}

	json.Unmarshal(jsonOutput, &updates)

	ts := time.Now().Unix()

	for _, update := range updates {
		n := update["name"].(string)
		pointsArray := update["points"].([]interface{})
		points := pointsArray[0].([]interface{})

		for index, column := range update["columns"].([]interface{}) {
			nn := n + "." + column.(string)
			v := points[index].(float64)

			out := fmt.Sprintf("%s %f %d", nn, v, ts)

			if *graphite == "" {
				log.Print(out)
			} else {
				conn.WriteToUDP([]byte(out), addr)
			}
		}
	}

	return nil
}

func main() {

	log.Print("Startup sysstatsd")

	// Determine hostname
	hostname, hosterr := os.Hostname()
	if hosterr != nil {
		log.Fatal("Unable to determine hostname")
	}

	// Parse and setup
	kingpin.Version("1.0.0")
	kingpin.Parse()

	if *showLoad == false && *showCPU == false && *showMemory == false && *filesystem == "" {
		log.Fatal("You must specify at least one of CPU, Load, Memory and Filesystem to output")
	}

	if *graphite != "" {
		connectToGraphite()
	} else {
		log.Print("No graphite server specified. Printing output to standard out.")
	}

	// System Statistics
	concreteSigar := sigar.ConcreteSigar{}
	cpuSamples, _ := concreteSigar.CollectCpuStats(500 * time.Millisecond)

	// Loop every interval
	for _ = range time.Tick(*interval) {

		updates := make([]interface{}, 0)

		// Load Averages
		if *showLoad {
			if avg, err := concreteSigar.GetLoadAverage(); err == nil {
				ls := LoadSeries{
					Name:    fmt.Sprintf("%v.load", hostname),
					Columns: []string{"one", "five", "fifteen"}}
				ls.Points[0] = []float64{
					roundLoad(avg.One),
					roundLoad(avg.Five),
					roundLoad(avg.Fifteen)}
				updates = append(updates, ls)

			} else {
				log.Fatal("Unable to get load average")
			}
		}

		// CPU
		if *showCPU {
			cpu := <-cpuSamples
			totalCpu := cpu.User + cpu.Nice + cpu.Sys + cpu.Idle + cpu.Wait + cpu.Irq + cpu.SoftIrq + cpu.Stolen
			cs := CPUSeries{
				Name:    fmt.Sprintf("%v.cpu", hostname),
				Columns: []string{"user", "sys", "wait", "idle"}}
			cs.Points[0] = []int{int(float64(cpu.User) / float64(totalCpu) * 100),
				int(float64(cpu.Sys) / float64(totalCpu) * 100),
				int(float64(cpu.Wait) / float64(totalCpu) * 100),
				int(float64(cpu.Idle) / float64(totalCpu) * 100)}
			updates = append(updates, cs)
		}

		// Memory
		if *showMemory {
			if mem, err := concreteSigar.GetMem(); err == nil {
				ms := MemSeries{
					Name:    fmt.Sprintf("%v.memory", hostname),
					Columns: []string{"percent"}}
				ms.Points[0] = []int{int(float64(mem.ActualUsed) / float64(mem.Total) * 100)}
				updates = append(updates, ms)

			} else {
				log.Fatal("Unable to determine memory statistics")
			}
		}

		// Disk
		if *filesystem != "" {
			if disk, err := concreteSigar.GetFileSystemUsage(*filesystem); err == nil {
				ds := DiskSeries{
					Name:    fmt.Sprintf("%v.disk", hostname),
					Columns: []string{"percent"}}
				ds.Points[0] = []int{int(float64(disk.Used) / float64(disk.Total) * 100)}
				updates = append(updates, ds)

			} else {
				log.Fatalf("Unable to get filesystem usage for %v", *filesystem)
			}
		}

		jsonOutput, err := json.Marshal(updates)
		if err != nil {
			log.Fatalf("Unabled to Marshal JSON: %v", err)
		}

		// Write to Graphite

		if err := writeData(jsonOutput); err != nil {
			log.Fatalf("Error writing to DB: %v", err)
		}
	}
}
