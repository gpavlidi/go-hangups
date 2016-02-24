package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/gpavlidi/go-hangups"
)

func OrDie(e error) {
	if e != nil {
		log.Fatal(e)
	}
}

func main() {
	s := &hangups.Session{}
	OrDie(s.Init())

	c := &hangups.Client{Session: s}

	// find whoami and seed the sync timestamp to current time
	getSelfInfo, _ := c.GetSelfInfo()
	serverNowUsecs := *getSelfInfo.ResponseHeader.CurrentServerTime

	ticker := time.NewTicker(time.Second * 5)
	for _ = range ticker.C {
		newEvents, _ := c.SyncAllNewEvents(serverNowUsecs, 1048576) //1 MB
		serverNowUsecs = *newEvents.ResponseHeader.CurrentServerTime

		for _, conversation := range newEvents.ConversationState {

			//find or generate conversation name
			conversationName := ""
			if conversation.Conversation.Name != nil {
				conversationName = fmt.Sprintf("hangouts-%s", *conversation.Conversation.Name)
			} else {
				participants := make([]string, 0)
				for _, participant := range conversation.Conversation.ParticipantData {
					// skip my name from participants list
					if *participant.Id.GaiaId == *getSelfInfo.SelfEntity.Id.GaiaId {
						continue
					}
					participants = append(participants, *participant.FallbackName)
				}
				conversationName = fmt.Sprintf("hangouts-%s", strings.Join(participants, ","))
			}

			for _, event := range conversation.Event {
				senderId := *event.SenderId.GaiaId

				// dont echo my msgs
				if senderId == *getSelfInfo.SelfEntity.Id.GaiaId {
					continue
				}

				// find sender name
				senderName := "Unknown"
				for _, participant := range conversation.Conversation.ParticipantData {
					if *participant.Id.GaiaId == senderId {
						senderName = *participant.FallbackName
						break
					}
				}

				// reconstruct msg text
				for _, segment := range event.ChatMessage.MessageContent.Segment {
					if *segment.Type != 1 {
						//skip SegmentType_SEGMENT_TYPE_LINE_BREAK
						fmt.Println("[", conversationName, "] ", senderName, ":", *segment.Text, " (", *segment.Type, ")")
					}
				}
			}

			// mark all events in this conversation as read
			_, _ = c.UpdateWatermark(*conversation.ConversationId.Id, serverNowUsecs)
		}
	}
}
