// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package mocks

import (
	"context"
	"sync"
	"time"

	"github.com/MainfluxLabs/mainflux/pkg/errors"
	"github.com/MainfluxLabs/mainflux/things"
)

var _ things.GroupRepository = (*groupRepositoryMock)(nil)

type groupRepositoryMock struct {
	mu sync.Mutex
	// Map of groups, group id as a key.
	// groups      map[GroupID]auth.Group
	groups map[string]things.Group
	// Map of group thing membership where thing id is a key and group id is a value.
	thingMembership map[string]string
	// Map of group thing where group id is a key and thing ids are values.
	things map[string][]string
	// Map of group channel membership where channel id is a key and group id is a value.
	channelMembership map[string]string
	// Map of group channel where group id is a key and channel ids are values.
	channels map[string][]string
}

// NewGroupRepository creates in-memory user repository
func NewGroupRepository() things.GroupRepository {
	return &groupRepositoryMock{
		groups:            make(map[string]things.Group),
		thingMembership:   make(map[string]string),
		things:            make(map[string][]string),
		channelMembership: make(map[string]string),
		channels:          make(map[string][]string),
	}
}

func (grm *groupRepositoryMock) Save(ctx context.Context, group things.Group) (things.Group, error) {
	grm.mu.Lock()
	defer grm.mu.Unlock()
	if _, ok := grm.groups[group.ID]; ok {
		return things.Group{}, errors.ErrConflict
	}

	grm.groups[group.ID] = group
	return group, nil
}

func (grm *groupRepositoryMock) Update(ctx context.Context, group things.Group) (things.Group, error) {
	grm.mu.Lock()
	defer grm.mu.Unlock()
	up, ok := grm.groups[group.ID]
	if !ok {
		return things.Group{}, errors.ErrNotFound
	}
	up.Name = group.Name
	up.Description = group.Description
	up.Metadata = group.Metadata
	up.UpdatedAt = time.Now()

	grm.groups[group.ID] = up
	return up, nil
}

func (grm *groupRepositoryMock) Remove(ctx context.Context, ids ...string) error {
	grm.mu.Lock()
	defer grm.mu.Unlock()

	for _, id := range ids {
		if _, ok := grm.groups[id]; !ok {
			return errors.ErrNotFound
		}

		for _, thingID := range grm.things[id] {
			delete(grm.thingMembership, thingID)
		}

		for _, channelID := range grm.channels[id] {
			delete(grm.channelMembership, channelID)
		}

		// This is not quite exact, it should go in depth
		delete(grm.groups, id)
	}
	return nil

}

func (grm *groupRepositoryMock) RetrieveAll(ctx context.Context) ([]things.Group, error) {
	grm.mu.Lock()
	defer grm.mu.Unlock()

	var items []things.Group
	for _, g := range grm.groups {
		items = append(items, g)
	}

	return items, nil
}

func (grm *groupRepositoryMock) RetrieveByID(ctx context.Context, id string) (things.Group, error) {
	grm.mu.Lock()
	defer grm.mu.Unlock()

	val, ok := grm.groups[id]
	if !ok {
		return things.Group{}, errors.ErrNotFound
	}
	return val, nil
}

func (grm *groupRepositoryMock) RetrieveByIDs(ctx context.Context, groupIDs []string) (things.GroupPage, error) {
	panic("not implemented")
}

func (grm *groupRepositoryMock) RetrieveByOwner(ctx context.Context, ownerID string, pm things.PageMetadata) (things.GroupPage, error) {
	grm.mu.Lock()
	defer grm.mu.Unlock()
	var items []things.Group
	for _, g := range grm.groups {
		items = append(items, g)
	}
	return things.GroupPage{
		Groups: items,
		PageMetadata: things.PageMetadata{
			Total: uint64(len(items)),
		},
	}, nil
}

func (grm *groupRepositoryMock) UnassignThing(ctx context.Context, groupID string, thingIDs ...string) error {
	grm.mu.Lock()
	defer grm.mu.Unlock()
	if _, ok := grm.groups[groupID]; !ok {
		return errors.ErrNotFound
	}

	for _, thingID := range thingIDs {
		things, ok := grm.things[groupID]
		if !ok {
			return errors.ErrNotFound
		}

		for i, th := range things {
			if th == thingID {
				grm.things[groupID] = append(things[:i], things[i+1:]...)
				delete(grm.thingMembership, thingID)
				break
			}
		}
	}

	return nil
}

func (grm *groupRepositoryMock) AssignThing(ctx context.Context, groupID string, thingIDs ...string) error {
	grm.mu.Lock()
	defer grm.mu.Unlock()

	if _, ok := grm.groups[groupID]; !ok {
		return errors.ErrNotFound
	}

	if _, ok := grm.things[groupID]; !ok {
		grm.things[groupID] = []string{}
	}

	for _, thingID := range thingIDs {
		grm.things[groupID] = append(grm.things[groupID], thingID)
		grm.thingMembership[thingID] = groupID
	}

	return nil
}

func (grm *groupRepositoryMock) RetrieveThingMembership(ctx context.Context, thingID string) (string, error) {
	grm.mu.Lock()
	defer grm.mu.Unlock()

	groupID, ok := grm.thingMembership[thingID]
	if !ok {
		return "", errors.ErrNotFound
	}
	return groupID, nil
}

func (grm *groupRepositoryMock) RetrieveGroupThings(ctx context.Context, ownerID, groupID string, pm things.PageMetadata) (things.GroupThingsPage, error) {
	grm.mu.Lock()
	defer grm.mu.Unlock()
	var items []things.Thing
	ths, ok := grm.things[groupID]
	if !ok {
		return things.GroupThingsPage{}, errors.ErrNotFound
	}

	first := uint64(pm.Offset)
	last := first + uint64(pm.Limit)

	if last > uint64(len(ths)) {
		last = uint64(len(ths))
	}

	for i := first; i < last; i++ {
		items = append(items, things.Thing{ID: ths[i]})
	}

	return things.GroupThingsPage{
		Things: items,
		PageMetadata: things.PageMetadata{
			Total: uint64(len(items)),
		},
	}, nil
}

func (grm *groupRepositoryMock) RetrieveAllThingRelations(ctx context.Context) ([]things.GroupThingRelation, error) {
	grm.mu.Lock()
	defer grm.mu.Unlock()

	var gtr []things.GroupThingRelation
	for grID, thIDs := range grm.things {
		for _, thID := range thIDs {
			gtr = append(gtr, things.GroupThingRelation{
				GroupID: grID,
				ThingID: thID,
			})
		}
	}

	return gtr, nil
}

func (grm *groupRepositoryMock) RetrieveAllChannelRelations(ctx context.Context) ([]things.GroupChannelRelation, error) {
	grm.mu.Lock()
	defer grm.mu.Unlock()

	var gcr []things.GroupChannelRelation
	for grID, chIDs := range grm.channels {
		for _, chID := range chIDs {
			gcr = append(gcr, things.GroupChannelRelation{
				GroupID:   grID,
				ChannelID: chID,
			})
		}
	}

	return gcr, nil
}

func (grm *groupRepositoryMock) AssignChannel(ctx context.Context, groupID string, channelIDs ...string) error {
	grm.mu.Lock()
	defer grm.mu.Unlock()

	if _, ok := grm.groups[groupID]; !ok {
		return errors.ErrNotFound
	}

	for _, channelID := range channelIDs {
		grm.channels[groupID] = append(grm.channels[groupID], channelID)
		grm.channelMembership[channelID] = groupID
	}

	return nil
}

func (grm *groupRepositoryMock) UnassignChannel(ctx context.Context, groupID string, channelIDs ...string) error {
	grm.mu.Lock()
	defer grm.mu.Unlock()

	if _, ok := grm.groups[groupID]; !ok {
		return errors.ErrNotFound
	}

	for _, channelID := range channelIDs {
		channels, ok := grm.channels[groupID]
		if !ok {
			return errors.ErrNotFound
		}

		for i, ch := range channels {
			if ch == channelID {
				grm.channels[groupID] = append(channels[:i], channels[i+1:]...)
				delete(grm.channelMembership, channelID)
				break
			}
		}
	}

	return nil
}

func (grm *groupRepositoryMock) RetrieveChannelMembership(ctx context.Context, channelID string) (string, error) {
	grm.mu.Lock()
	defer grm.mu.Unlock()

	groupID, ok := grm.channelMembership[channelID]
	if !ok {
		return "", errors.ErrNotFound
	}

	return groupID, nil
}

func (grm *groupRepositoryMock) RetrieveGroupChannels(ctx context.Context, ownerID, groupID string, pm things.PageMetadata) (things.GroupChannelsPage, error) {
	grm.mu.Lock()
	defer grm.mu.Unlock()

	var items []things.Channel
	chs, ok := grm.channels[groupID]
	if !ok {
		return things.GroupChannelsPage{}, nil
	}

	first := uint64(pm.Offset)
	last := first + uint64(pm.Limit)

	if last > uint64(len(chs)) {
		last = uint64(len(chs))
	}

	for i := first; i < last; i++ {
		items = append(items, things.Channel{ID: chs[i]})
	}

	return things.GroupChannelsPage{
		Channels: items,
		PageMetadata: things.PageMetadata{
			Total: uint64(len(items)),
		},
	}, nil
}

func (grm *groupRepositoryMock) RetrieveGroupThingsByChannel(ctx context.Context, groupID, channelID string, pm things.PageMetadata) (things.GroupThingsPage, error) {
	panic("not implemented")
}

func (grm *groupRepositoryMock) RetrieveByAdmin(ctx context.Context, pm things.PageMetadata) (things.GroupPage, error) {
	panic("not implemented")
}
