package helper

import "fmt"

type GroupMemberId struct {
	ObjectSubResourceId
	GroupId  string
	MemberId string
}

func GroupMemberIdFrom(groupId, memberId string) GroupMemberId {
	return GroupMemberId{
		ObjectSubResourceId: ObjectSubResourceIdFrom(groupId, "member", memberId),
		GroupId:             groupId,
		MemberId:            memberId,
	}
}

func ParseGroupMemberId(idString string) (*GroupMemberId, error) {
	id, err := ParseObjectSubResourceId(idString, "member")
	if err != nil {
		return nil, fmt.Errorf("unable to parse Member ID: %v", err)
	}

	return &GroupMemberId{
		ObjectSubResourceId: *id,
		GroupId:             id.objectId,
		MemberId:            id.subId,
	}, nil
}
