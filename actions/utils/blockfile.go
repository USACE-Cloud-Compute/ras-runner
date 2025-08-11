package utils

import (
	"encoding/json"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
)

type RealzBlock struct {
	RealzIndex       int   `json:"realization_index"`
	BlockIndex       int   `json:"block_index"`
	BlockEventClount int   `json:"block_event_count"`
	BlockEventStart  int   `json:"block_event_start"`
	BlockEventEnd    int   `json:"block_event_end"`
	ImportantEvents  []int `json:"-"`
}

type BlockfileInput struct {
	BlockFilePath            string
	ImportantEventsFilePath  string
	StartRealz               int
	EndRealz                 int
	ProcessImpEventsPerBlock bool
}

type BlockFile struct {
	input           BlockfileInput
	blocks          []RealzBlock
	importantEvents []int
}

func NewBlockFile(input BlockfileInput) (BlockFile, error) {

	blockfile := BlockFile{}

	blockReader, err := os.Open(input.BlockFilePath)
	if err != nil {
		return blockfile, err
	}

	blockfile.blocks = []RealzBlock{}
	err = json.NewDecoder(blockReader).Decode(&blockfile.blocks)
	if err != nil {
		return blockfile, err
	}

	//important events are optional
	if input.ImportantEventsFilePath != "" {
		blockfile.loadImportantEvents(input.ImportantEventsFilePath)
	}

	//get realz range if necessary
	if input.StartRealz == 0 || input.EndRealz == 0 {
		minRealz, maxRealz := blockfile.getRealzRange()
		input.StartRealz = minRealz
		input.EndRealz = maxRealz
	}

	if input.ProcessImpEventsPerBlock {
		blockfile.processBlockImpEvents(input.StartRealz, input.EndRealz)
	}

	// if input.ProcessImpEventsPerRealz {
	// 	blockfile.processEventsPerRealz()
	// }

	blockfile.input = input

	return blockfile, nil
}

func (bf *BlockFile) loadImportantEvents(importatEventsFilePath string) error {

	importantEventsReader, err := os.Open(importatEventsFilePath)
	if err != nil {
		return err
	}

	importantEventsContent, err := io.ReadAll(importantEventsReader)
	if err != nil {
		return err
	}

	importantEventStrings := strings.Split(string(importantEventsContent), ",")
	if err != nil {
		return err
	}
	bf.importantEvents = importantEventsToInt(importantEventStrings)

	return nil
}

func importantEventsToInt(importantEvents []string) []int {
	intEvents := make([]int, len(importantEvents))
	for i := 0; i < len(importantEvents); i++ {
		newEvent, err := strconv.Atoi(strings.TrimSpace(importantEvents[i]))
		if err != nil {
			log.Printf("Unable to convert %s to int. ...Skipping\n", importantEvents[i])
			continue
		}
		intEvents[i] = newEvent
	}
	return intEvents
}

func (bf *BlockFile) processBlockImpEvents(startRealz int, endRealz int) map[int][]RealzBlock {
	realizations := make(map[int][]RealzBlock)
	for b := 0; b < len(bf.blocks); b++ {
		if bf.blocks[b].RealzIndex >= startRealz && bf.blocks[b].RealzIndex <= endRealz { //if so is this in the processing event range
			for i := 0; i < len(bf.importantEvents); i++ {
				if bf.importantEvents[i] >= bf.blocks[b].BlockEventStart && bf.importantEvents[i] <= bf.blocks[b].BlockEventEnd {
					//append important events to block
					if bf.blocks[b].ImportantEvents == nil {
						bf.blocks[b].ImportantEvents = []int{bf.importantEvents[i]}
					} else {
						bf.blocks[b].ImportantEvents = append(bf.blocks[b].ImportantEvents, bf.importantEvents[i])
					}
				}
			}
			//add block to realizations data structure
			if realz, ok := realizations[bf.blocks[b].RealzIndex]; ok {
				realizations[bf.blocks[b].RealzIndex] = append(realz, bf.blocks[b])
			} else {
				realizations[bf.blocks[b].RealzIndex] = []RealzBlock{bf.blocks[b]}
			}
		}
	}
	return realizations
}

func (bf *BlockFile) getRealzRange() (min int, max int) {

	maxRealz := -1
	minRealz := -1
	for _, rg := range bf.blocks {
		if maxRealz < rg.RealzIndex {
			maxRealz = rg.RealzIndex
		}
		if minRealz == -1 || minRealz > rg.RealzIndex {
			minRealz = rg.RealzIndex
		}
	}
	return minRealz, maxRealz
}

func (bf *BlockFile) getRealzEventRange(realz int) (start int, end int) {

	startEvent := 0
	endEvent := 0

	for _, b := range bf.blocks {
		if b.RealzIndex == realz {
			if startEvent == 0 || startEvent > b.BlockEventStart {
				startEvent = b.BlockEventStart
			}
			if b.BlockEventEnd > endEvent {
				endEvent = b.BlockEventEnd
			}
		}
	}
	return startEvent, endEvent
}
