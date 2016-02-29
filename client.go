package hangups

import (
	"encoding/base64"
	"errors"
	"fmt"
	"math/rand"
	"net/url"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/gpavlidi/go-hangups/proto"
)

type Client struct {
	Session  *Session
	ClientId string
}

// initialize random number generator needed for client id
func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}

/*
* Helper Methods
*
*
 */
func (c *Client) NewRequestHeaders() *hangouts.RequestHeader {
	version := "hangups-0.0.1"
	language := "en"

	return &hangouts.RequestHeader{
		ClientVersion:    &hangouts.ClientVersion{MajorVersion: &version},
		ClientIdentifier: &hangouts.ClientIdentifier{Resource: nil}, //use request_header.client_identifier.resource
		LanguageCode:     &language,
	}
}

func (c *Client) NewEventRequestHeaders(conversationId string, offTheRecord bool) *hangouts.EventRequestHeader {
	expectedOtr := hangouts.OffTheRecordStatus_OFF_THE_RECORD_STATUS_ON_THE_RECORD
	if offTheRecord {
		expectedOtr = hangouts.OffTheRecordStatus_OFF_THE_RECORD_STATUS_OFF_THE_RECORD
	}
	deliveryMedium := hangouts.DeliveryMediumType_DELIVERY_MEDIUM_BABEL
	// needs to be unique every time
	clientGeneratedId := uint64(rand.Uint32())
	eventType := hangouts.EventType_EVENT_TYPE_REGULAR_CHAT_MESSAGE
	return &hangouts.EventRequestHeader{
		ConversationId:    &hangouts.ConversationId{Id: &conversationId},
		ClientGeneratedId: &clientGeneratedId,
		ExpectedOtr:       &expectedOtr,
		DeliveryMedium:    &hangouts.DeliveryMedium{MediumType: &deliveryMedium},
		EventType:         &eventType,
	}
}

func (c *Client) ProtobufApiRequest(apiEndpoint string, requestStruct, responseStruct proto.Message) error {
	payload, err := proto.Marshal(requestStruct)
	if err != nil {
		return err
	}
	url := fmt.Sprintf("https://clients6.google.com/chat/v1/%s", apiEndpoint)
	output, err := ApiRequest(url, "application/x-protobuf", "proto", c.Session.Cookies, c.Session.Sapisid, map[string]string{}, payload)
	if err != nil {
		return err
	}

	decodedOutput, err := base64.StdEncoding.DecodeString(string(output))
	if err != nil {
		return err
	}

	// if debugging
	//err = ioutil.WriteFile("./proto.bin", decodedOutput, 0644)
	// and then `protoc --decode_raw < proto.bin`

	err = proto.Unmarshal(decodedOutput, responseStruct)
	if err != nil {
		return err
	}

	return nil
}

/*
* Api WrapperMethods
*
*
 */

//Invite users to join an existing group conversation.
func (c *Client) AddUser(inviteesGaiaIds []string, conversationId string) (*hangouts.AddUserResponse, error) {
	inviteeIds := make([]*hangouts.InviteeID, len(inviteesGaiaIds))
	for ind, gaiaId := range inviteesGaiaIds {
		inviteeIds[ind] = &hangouts.InviteeID{GaiaId: &gaiaId}
	}

	request := &hangouts.AddUserRequest{
		RequestHeader:      c.NewRequestHeaders(),
		InviteeId:          inviteeIds,
		EventRequestHeader: c.NewEventRequestHeaders(conversationId, false),
	}
	response := &hangouts.AddUserResponse{}
	err := c.ProtobufApiRequest("conversations/adduser", request, response)
	if err != nil {
		return nil, err
	}
	return response, nil
}

//Create a new conversation.
func (c *Client) CreateConversation(inviteesGaiaIds []string, name string, oneOnOne bool) (*hangouts.CreateConversationResponse, error) {
	inviteeIds := make([]*hangouts.InviteeID, len(inviteesGaiaIds))
	for ind, gaiaId := range inviteesGaiaIds {
		inviteeIds[ind] = &hangouts.InviteeID{GaiaId: &gaiaId}
	}

	conversationType := hangouts.ConversationType_CONVERSATION_TYPE_GROUP
	if oneOnOne {
		conversationType = hangouts.ConversationType_CONVERSATION_TYPE_ONE_TO_ONE
	}
	clientGeneratedId := uint64(rand.Uint32())
	request := &hangouts.CreateConversationRequest{
		RequestHeader:     c.NewRequestHeaders(),
		InviteeId:         inviteeIds,
		ClientGeneratedId: &clientGeneratedId,
		Name:              &name,
		Type:              &conversationType,
	}
	response := &hangouts.CreateConversationResponse{}
	err := c.ProtobufApiRequest("conversations/createconversation", request, response)
	if err != nil {
		return nil, err
	}
	return response, nil
}

/*
Leave a one-to-one conversation.
One-to-one conversations are "sticky"; they can't actually be deleted.
This API clears the event history of the specified conversation up to
delete_upper_bound_timestamp, hiding it if no events remain.
*/
func (c *Client) DeleteConversation(conversationId string, deleteUpperBoundTimestamp uint64) (*hangouts.DeleteConversationResponse, error) {
	request := &hangouts.DeleteConversationRequest{
		RequestHeader:             c.NewRequestHeaders(),
		ConversationId:            &hangouts.ConversationId{Id: &conversationId},
		DeleteUpperBoundTimestamp: &deleteUpperBoundTimestamp,
	}
	response := &hangouts.DeleteConversationResponse{}
	err := c.ProtobufApiRequest("conversations/deleteconversation", request, response)
	if err != nil {
		return nil, err
	}
	return response, nil
}

// Send an easter egg event to a conversation.
func (c *Client) EasterEgg(conversationId string, message string) (*hangouts.EasterEggResponse, error) {
	request := &hangouts.EasterEggRequest{
		RequestHeader:  c.NewRequestHeaders(),
		ConversationId: &hangouts.ConversationId{Id: &conversationId},
		EasterEgg:      &hangouts.EasterEgg{Message: &message},
	}
	response := &hangouts.EasterEggResponse{}
	err := c.ProtobufApiRequest("conversations/easteregg", request, response)
	if err != nil {
		return nil, err
	}
	return response, nil
}

// Return conversation info and recent events.
func (c *Client) GetConversation(conversationId string, includeEvent bool, maxEventsPerConversation uint64) (*hangouts.GetConversationResponse, error) {
	request := &hangouts.GetConversationRequest{
		RequestHeader:            c.NewRequestHeaders(),
		ConversationSpec:         &hangouts.ConversationSpec{ConversationId: &hangouts.ConversationId{Id: &conversationId}},
		IncludeEvent:             &includeEvent,
		MaxEventsPerConversation: &maxEventsPerConversation,
	}
	response := &hangouts.GetConversationResponse{}
	err := c.ProtobufApiRequest("conversations/getconversation", request, response)
	if err != nil {
		return nil, err
	}
	return response, nil
}

// Return info about a list of users.
func (c *Client) GetEntityById(gaiaIds []string) (*hangouts.GetEntityByIdResponse, error) {
	batchLookupSpec := make([]*hangouts.EntityLookupSpec, len(gaiaIds))
	for ind, gaiaId := range gaiaIds {
		batchLookupSpec[ind] = &hangouts.EntityLookupSpec{GaiaId: &gaiaId}
	}
	request := &hangouts.GetEntityByIdRequest{
		RequestHeader:   c.NewRequestHeaders(),
		BatchLookupSpec: batchLookupSpec,
	}
	response := &hangouts.GetEntityByIdResponse{}
	err := c.ProtobufApiRequest("contacts/getentitybyid", request, response)
	if err != nil {
		return nil, err
	}
	return response, nil
}

// Return info about the current user.
func (c *Client) GetSelfInfo() (*hangouts.GetSelfInfoResponse, error) {
	request := &hangouts.GetSelfInfoRequest{
		RequestHeader: c.NewRequestHeaders(),
	}
	response := &hangouts.GetSelfInfoResponse{}
	err := c.ProtobufApiRequest("contacts/getselfinfo", request, response)
	if err != nil {
		return nil, err
	}
	return response, nil
}

/*
	Return presence status for a list of users.
	doesnt support passing an array of gaiaIds.
	fails with:
	{"status":4,"error_description":"Duplicate ParticipantIds in request"}
*/
func (c *Client) QueryPresence(gaiaId string) (*hangouts.QueryPresenceResponse, error) {
	request := &hangouts.QueryPresenceRequest{
		RequestHeader: c.NewRequestHeaders(),
		ParticipantId: []*hangouts.ParticipantId{&hangouts.ParticipantId{GaiaId: &gaiaId, ChatId: &gaiaId}},
		FieldMask:     []hangouts.FieldMask{hangouts.FieldMask_FIELD_MASK_REACHABLE, hangouts.FieldMask_FIELD_MASK_AVAILABLE, hangouts.FieldMask_FIELD_MASK_DEVICE},
	}
	response := &hangouts.QueryPresenceResponse{}
	err := c.ProtobufApiRequest("presence/querypresence", request, response)
	if err != nil {
		return nil, err
	}
	return response, nil
}

// Leave a group conversation.
func (c *Client) RemoveUser(conversationId string) (*hangouts.RemoveUserResponse, error) {
	request := &hangouts.RemoveUserRequest{
		RequestHeader:      c.NewRequestHeaders(),
		EventRequestHeader: c.NewEventRequestHeaders(conversationId, false),
	}
	response := &hangouts.RemoveUserResponse{}
	err := c.ProtobufApiRequest("conversations/removeuser", request, response)
	if err != nil {
		return nil, err
	}
	return response, nil
}

/*
	Rename a conversation.
    Both group and one-to-one conversations may be renamed, but the
    official Hangouts clients have mixed support for one-to-one
    conversations with custom names.
*/
func (c *Client) RenameConversation(conversationId, newName string) (*hangouts.RenameConversationResponse, error) {
	request := &hangouts.RenameConversationRequest{
		RequestHeader:      c.NewRequestHeaders(),
		EventRequestHeader: c.NewEventRequestHeaders(conversationId, false),
		NewName:            &newName,
	}
	response := &hangouts.RenameConversationResponse{}
	err := c.ProtobufApiRequest("conversations/renameconversation", request, response)
	if err != nil {
		return nil, err
	}
	return response, nil
}

// Return info for users based on a query.
func (c *Client) SearchEntities(query string, maxCount uint64) (*hangouts.SearchEntitiesResponse, error) {
	request := &hangouts.SearchEntitiesRequest{
		RequestHeader: c.NewRequestHeaders(),
		Query:         &query,
		MaxCount:      &maxCount,
	}
	response := &hangouts.SearchEntitiesResponse{}
	err := c.ProtobufApiRequest("contacts/searchentities", request, response)
	if err != nil {
		return nil, err
	}
	return response, nil
}

// Send a chat message to a conversation.
func (c *Client) SendChatMessage(conversationId, message string) (*hangouts.SendChatMessageResponse, error) {
	// for now treat message as one segment
	// TODO split it on TEXT, LINE_BREAK and LINK
	segmentType := hangouts.SegmentType_SEGMENT_TYPE_TEXT

	linkData := &hangouts.LinkData{}
	// check if it is a link
	_, err := url.Parse(message)
	if err == nil {
		segmentType = hangouts.SegmentType_SEGMENT_TYPE_LINK
		linkData.LinkTarget = &message
	}

	messageContent := &hangouts.MessageContent{
		Segment: []*hangouts.Segment{
			&hangouts.Segment{
				Type:       &segmentType,
				Text:       &message,
				Formatting: &hangouts.Formatting{},
				LinkData:   linkData,
			},
		},
		Attachment: nil,
	}

	request := &hangouts.SendChatMessageRequest{
		RequestHeader:      c.NewRequestHeaders(),
		EventRequestHeader: c.NewEventRequestHeaders(conversationId, false),
		MessageContent:     messageContent,
		//Annotation: [],
		//ExistingMedia: &hangouts.ExistingMedia{Photo: &hangouts.Photo{}}, //picassa photos
	}
	response := &hangouts.SendChatMessageResponse{}
	err = c.ProtobufApiRequest("conversations/sendchatmessage", request, response)
	if err != nil {
		return nil, err
	}
	return response, nil
}

// Send an invitation to a non-contact.
func (c *Client) SendOffnetworkInvitation(email string) (*hangouts.SendOffnetworkInvitationResponse, error) {
	addressType := hangouts.OffnetworkAddressType_OFFNETWORK_ADDRESS_TYPE_EMAIL
	request := &hangouts.SendOffnetworkInvitationRequest{
		RequestHeader: c.NewRequestHeaders(),
		InviteeAddress: &hangouts.OffnetworkAddress{
			Type:  &addressType,
			Email: &email,
		},
	}
	response := &hangouts.SendOffnetworkInvitationResponse{}
	err := c.ProtobufApiRequest("devices/sendoffnetworkinvitation", request, response)
	if err != nil {
		return nil, err
	}
	return response, nil
}

// Set the active client.
// timeout is 120 secs in hangups
func (c *Client) SetActiveClient(email string, isActive bool, timeoutSecs uint64) (*hangouts.SetActiveClientResponse, error) {
	if c.ClientId == "" {
		return nil, errors.New("Can't set active client without a ClientId")
	}
	emailAndResource := fmt.Sprintf("%s/%s", email, c.ClientId)

	request := &hangouts.SetActiveClientRequest{
		RequestHeader: c.NewRequestHeaders(),
		IsActive:      &isActive,
		FullJid:       &emailAndResource,
		TimeoutSecs:   &timeoutSecs,
	}
	response := &hangouts.SetActiveClientResponse{}
	err := c.ProtobufApiRequest("clients/setactiveclient", request, response)
	if err != nil {
		return nil, err
	}
	return response, nil
}

// Set the notification level of a conversation.
func (c *Client) SetConversationNotificationLevel(conversationId string, setQuiet bool) (*hangouts.SetConversationNotificationLevelResponse, error) {
	notificationLevel := hangouts.NotificationLevel_NOTIFICATION_LEVEL_RING
	if setQuiet {
		notificationLevel = hangouts.NotificationLevel_NOTIFICATION_LEVEL_QUIET
	}
	request := &hangouts.SetConversationNotificationLevelRequest{
		RequestHeader:  c.NewRequestHeaders(),
		Level:          &notificationLevel,
		ConversationId: &hangouts.ConversationId{Id: &conversationId},
	}
	response := &hangouts.SetConversationNotificationLevelResponse{}
	err := c.ProtobufApiRequest("conversations/setconversationnotificationlevel", request, response)
	if err != nil {
		return nil, err
	}
	return response, nil
}

// Set focus to a conversation.
func (c *Client) SetFocus(conversationId string, unfocus bool, timeoutSecs uint32) (*hangouts.SetFocusResponse, error) {
	isFocused := hangouts.FocusType_FOCUS_TYPE_FOCUSED
	if unfocus {
		isFocused = hangouts.FocusType_FOCUS_TYPE_UNFOCUSED
	}
	request := &hangouts.SetFocusRequest{
		RequestHeader:  c.NewRequestHeaders(),
		ConversationId: &hangouts.ConversationId{Id: &conversationId},
		Type:           &isFocused,
		TimeoutSecs:    &timeoutSecs,
	}
	response := &hangouts.SetFocusResponse{}
	err := c.ProtobufApiRequest("conversations/setfocus", request, response)
	if err != nil {
		return nil, err
	}
	return response, nil
}

// Set the presence status.
// presenceState: 1=NONE 30=IDLE 40=ACTIVE
func (c *Client) SetPresence(presenceState int32, timeoutSecs uint64) (*hangouts.SetPresenceResponse, error) {
	request := &hangouts.SetPresenceRequest{
		RequestHeader: c.NewRequestHeaders(),
		PresenceStateSetting: &hangouts.PresenceStateSetting{
			TimeoutSecs: &timeoutSecs,
			Type:        (*hangouts.ClientPresenceStateType)(&presenceState),
		},
	}
	response := &hangouts.SetPresenceResponse{}
	err := c.ProtobufApiRequest("conversations/setfocus", request, response)
	if err != nil {
		return nil, err
	}
	return response, nil
}

// Set the typing status of a conversation.
// typingState: 1=Started 2=Paused 3=Stopped
func (c *Client) SetTyping(conversationId string, typingState int32) (*hangouts.SetTypingResponse, error) {
	request := &hangouts.SetTypingRequest{
		RequestHeader:  c.NewRequestHeaders(),
		ConversationId: &hangouts.ConversationId{Id: &conversationId},
		Type:           (*hangouts.TypingType)(&typingState),
	}
	response := &hangouts.SetTypingResponse{}
	err := c.ProtobufApiRequest("conversations/settyping", request, response)
	if err != nil {
		return nil, err
	}
	return response, nil
}

// List all events occurring at or after a timestamp.
func (c *Client) SyncAllNewEvents(lastSyncTimestamp, maxResponseSizeBytes uint64) (*hangouts.SyncAllNewEventsResponse, error) {
	request := &hangouts.SyncAllNewEventsRequest{
		RequestHeader:        c.NewRequestHeaders(),
		LastSyncTimestamp:    &lastSyncTimestamp,
		MaxResponseSizeBytes: &maxResponseSizeBytes,
	}
	response := &hangouts.SyncAllNewEventsResponse{}
	err := c.ProtobufApiRequest("conversations/syncallnewevents", request, response)
	if err != nil {
		return nil, err
	}
	return response, nil
}

// Return info on recent conversations and their events.
func (c *Client) SyncRecentConversations(maxConversations, maxEventsPerConversation uint64) (*hangouts.SyncRecentConversationsResponse, error) {
	request := &hangouts.SyncRecentConversationsRequest{
		RequestHeader:            c.NewRequestHeaders(),
		MaxConversations:         &maxConversations,
		MaxEventsPerConversation: &maxEventsPerConversation,
		SyncFilter:               []hangouts.SyncFilter{hangouts.SyncFilter_SYNC_FILTER_INBOX},
	}
	response := &hangouts.SyncRecentConversationsResponse{}
	err := c.ProtobufApiRequest("conversations/syncrecentconversations", request, response)
	if err != nil {
		return nil, err
	}
	return response, nil
}

// Update the watermark (read timestamp) of a conversation.
func (c *Client) UpdateWatermark(conversationId string, lastReadTimestamp uint64) (*hangouts.UpdateWatermarkResponse, error) {
	request := &hangouts.UpdateWatermarkRequest{
		RequestHeader:     c.NewRequestHeaders(),
		ConversationId:    &hangouts.ConversationId{Id: &conversationId},
		LastReadTimestamp: &lastReadTimestamp,
	}
	response := &hangouts.UpdateWatermarkResponse{}
	err := c.ProtobufApiRequest("conversations/updatewatermark", request, response)
	if err != nil {
		return nil, err
	}
	return response, nil
}
