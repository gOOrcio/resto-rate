package models

import (
	friendshippb "api/src/generated/friendship/v1"
	"time"

	"gorm.io/gorm"
)

const (
	FriendRequestStatusPending  = "pending"
	FriendRequestStatusAccepted = "accepted"
	FriendRequestStatusDeclined = "declined"
)

type FriendRequest struct {
	UUIDv7
	SenderID   string    `gorm:"not null;index"`
	ReceiverID string    `gorm:"not null;index"`
	// PairKey is a canonical, unordered representation of the sender/receiver pair
	// (min_id + ":" + max_id) to enforce uniqueness regardless of request direction.
	PairKey  string    `gorm:"not null;uniqueIndex"`
	Sender   User      `gorm:"foreignKey:SenderID"`
	Receiver User      `gorm:"foreignKey:ReceiverID"`
	Status   string    `gorm:"not null;default:'pending'"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

func (f *FriendRequest) BeforeCreate(tx *gorm.DB) (err error) {
	if err = f.UUIDv7.BeforeCreate(tx); err != nil {
		return err
	}
	if f.SenderID != "" && f.ReceiverID != "" {
		if f.SenderID < f.ReceiverID {
			f.PairKey = f.SenderID + ":" + f.ReceiverID
		} else {
			f.PairKey = f.ReceiverID + ":" + f.SenderID
		}
	}
	return nil
}

func (f *FriendRequest) ToProto() *friendshippb.FriendRequestProto {
	status := friendshippb.FriendRequestStatus_PENDING
	switch f.Status {
	case FriendRequestStatusAccepted:
		status = friendshippb.FriendRequestStatus_ACCEPTED
	case FriendRequestStatusDeclined:
		status = friendshippb.FriendRequestStatus_DECLINED
	}
	return &friendshippb.FriendRequestProto{
		Id:            f.ID,
		SenderId:      f.SenderID,
		ReceiverId:    f.ReceiverID,
		Status:        status,
		CreatedAt:     f.CreatedAt.Unix(),
		SenderName:    f.Sender.Name,
		ReceiverName:  f.Receiver.Name,
		SenderEmail:   derefString(f.Sender.Email),
		ReceiverEmail: derefString(f.Receiver.Email),
	}
}
