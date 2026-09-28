package api

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"

	"git.sr.ht/~rockorager/go-jmap"
	"git.sr.ht/~rockorager/go-jmap/mail/email"
	"git.sr.ht/~rockorager/go-jmap/mail/mailbox"
)

// MoveChunk caps the emails patched by one Email/set when a delete empties a
// mailbox.
const MoveChunk = 100

// CheckBoxName refuses an empty name, a name with a slash (paths are
// slash-joined, so it could never be told apart from a subfolder), and a name
// a sibling under parentID already has. except is the mailbox being renamed.
func CheckBoxName(boxes []Box, parentID, name, except string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return errors.New("the mailbox name is empty")
	}
	if strings.Contains(name, "/") {
		return fmt.Errorf("%q has a slash in it; to make a subfolder, pass the parent with --parent", name)
	}
	for _, b := range boxes {
		if b.ParentID == parentID && b.ID != except && strings.EqualFold(b.Name, name) {
			return fmt.Errorf("a mailbox named %q already exists there: %s (%s)", b.Name, b.Path, b.ID)
		}
	}
	return nil
}

// CheckDeletable is the delete guard, in order: a mailbox with a role is never
// deleted, nor one with subfolders, nor one that still holds email unless the
// email is to be moved out first.
func CheckDeletable(boxes []Box, box *Box, moving bool) error {
	if box.Kind != "" {
		return fmt.Errorf("%s has the %q role and cannot be deleted; Fastmail relies on it", box.Path, box.Kind)
	}
	var children []string
	for _, b := range boxes {
		if b.ParentID == box.ID {
			children = append(children, b.Path)
		}
	}
	if len(children) > 0 {
		sort.Strings(children)
		return fmt.Errorf("%s has %d subfolder%s (%s); delete or move them first", box.Path, len(children), pluralS(len(children)), strings.Join(children, ", "))
	}
	if box.TotalCount > 0 && !moving {
		return fmt.Errorf("%s holds %d email%s; pass --move-to <mailbox> (or --move-to alone for Trash) to move them out first", box.Path, box.TotalCount, pluralS(int(box.TotalCount)))
	}
	return nil
}

func pluralS(n int) string {
	if n == 1 {
		return ""
	}
	return "s"
}

// CreateBox creates a mailbox named name under parentID ("" for the top
// level) and returns its id.
func (c *Client) CreateBox(ctx context.Context, account, parentID, name string) (string, error) {
	req := &jmap.Request{}
	call := req.Invoke(&mailbox.Set{
		Account: jmap.ID(account),
		Create:  map[jmap.ID]*mailbox.Mailbox{"new": {Name: name, ParentID: jmap.ID(parentID)}},
	})
	set, err := c.mailboxSet(ctx, req, call)
	if err != nil {
		return "", err
	}
	if e, ok := set.NotCreated["new"]; ok {
		return "", fmt.Errorf("Fastmail did not create the mailbox: %s", setErrorText(e))
	}
	created, ok := set.Created["new"]
	if !ok || created == nil || created.ID == "" {
		return "", errors.New("Fastmail did not report the new mailbox")
	}
	return string(created.ID), nil
}

// RenameBox changes a mailbox's name, leaving its parent and children alone.
func (c *Client) RenameBox(ctx context.Context, account, id, name string) error {
	req := &jmap.Request{}
	call := req.Invoke(&mailbox.Set{
		Account: jmap.ID(account),
		Update:  map[jmap.ID]jmap.Patch{jmap.ID(id): {"name": name}},
	})
	set, err := c.mailboxSet(ctx, req, call)
	if err != nil {
		return err
	}
	if e, ok := set.NotUpdated[jmap.ID(id)]; ok {
		return fmt.Errorf("Fastmail did not rename the mailbox: %s", setErrorText(e))
	}
	if _, ok := set.Updated[jmap.ID(id)]; !ok {
		return errors.New("Fastmail did not confirm the rename")
	}
	return nil
}

// DestroyBox destroys an empty mailbox. It never asks the server to remove
// the mailbox's email: a mailbox that still holds any is refused by Fastmail.
func (c *Client) DestroyBox(ctx context.Context, account, id string) error {
	req := &jmap.Request{}
	call := req.Invoke(&mailbox.Set{
		Account:               jmap.ID(account),
		Destroy:               []jmap.ID{jmap.ID(id)},
		OnDestroyRemoveEmails: false,
	})
	set, err := c.mailboxSet(ctx, req, call)
	if err != nil {
		return err
	}
	if e, ok := set.NotDestroyed[jmap.ID(id)]; ok {
		return fmt.Errorf("Fastmail did not delete the mailbox: %s", setErrorText(e))
	}
	for _, d := range set.Destroyed {
		if d == jmap.ID(id) {
			return nil
		}
	}
	return errors.New("Fastmail did not confirm the delete")
}

// MoveAllEmails moves every email in the mailbox from into the mailbox to,
// MoveChunk emails per Email/set, and returns how many moved. It stops at the
// first chunk the server refuses in part.
func (c *Client) MoveAllEmails(ctx context.Context, account, from, to string) (int, error) {
	acct := jmap.ID(account)
	var ids []jmap.ID
	for {
		req := &jmap.Request{}
		call := req.Invoke(&email.Query{
			Account:  acct,
			Filter:   &email.FilterCondition{InMailbox: jmap.ID(from)},
			Sort:     []*email.SortComparator{{Property: "receivedAt", IsAscending: false}},
			Position: int64(len(ids)),
			Limit:    MoveChunk,
		})
		resp, err := c.do(ctx, req)
		if err != nil {
			return 0, err
		}
		args, err := responseFor(resp, call)
		if err != nil {
			return 0, err
		}
		got, ok := args.(*email.QueryResponse)
		if !ok {
			return 0, errors.New("unexpected Email/query response")
		}
		ids = append(ids, got.IDs...)
		if len(got.IDs) < MoveChunk {
			break
		}
	}

	moved := 0
	for start := 0; start < len(ids); start += MoveChunk {
		end := min(start+MoveChunk, len(ids))
		update := make(map[jmap.ID]jmap.Patch, end-start)
		for _, id := range ids[start:end] {
			update[id] = jmap.Patch{"mailboxIds/" + to: true, "mailboxIds/" + from: nil}
		}
		req := &jmap.Request{}
		call := req.Invoke(&email.Set{Account: acct, Update: update})
		resp, err := c.do(ctx, req)
		if err != nil {
			return moved, err
		}
		args, err := responseFor(resp, call)
		if err != nil {
			return moved, err
		}
		set, ok := args.(*email.SetResponse)
		if !ok {
			return moved, errors.New("unexpected Email/set response")
		}
		moved += len(set.Updated)
		if len(set.NotUpdated) > 0 {
			var failed []string
			for id, e := range set.NotUpdated {
				failed = append(failed, fmt.Sprintf("%s (%s)", id, setErrorText(e)))
			}
			sort.Strings(failed)
			return moved, fmt.Errorf("could not move %d email%s: %s", len(failed), pluralS(len(failed)), strings.Join(failed, ", "))
		}
	}
	return moved, nil
}

func (c *Client) mailboxSet(ctx context.Context, req *jmap.Request, call string) (*mailbox.SetResponse, error) {
	resp, err := c.do(ctx, req)
	if err != nil {
		return nil, err
	}
	args, err := responseFor(resp, call)
	if err != nil {
		return nil, err
	}
	set, ok := args.(*mailbox.SetResponse)
	if !ok {
		return nil, errors.New("unexpected Mailbox/set response")
	}
	return set, nil
}

func setErrorText(e *jmap.SetError) string {
	if e == nil {
		return "unknown error"
	}
	if e.Description != nil && *e.Description != "" {
		return e.Type + ": " + *e.Description
	}
	return e.Type
}
