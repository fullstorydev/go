package chatterbox

import (
	"sort"
	"strings"
)

// MembersModel is a basic model object representing the current users in the room.
// Used by both client and server.  For this app, we've made a couple of design choices:
// 1) Copy-on-read instead of copy-on-write.
// 2) Simple direct mutators (Add/Remove).
// Contrast with MembersModelAlt.
type MembersModel map[string]struct{}

func (mm MembersModel) Add(name string) {
	mm[name] = struct{}{}
}

func (mm MembersModel) Remove(name string) {
	delete(mm, name)
}

// Strings returns a copy of the current list of members.
func (mm MembersModel) Strings() []string {
	var ret []string
	for k := range mm {
		ret = append(ret, k)
	}
	sort.Strings(ret)
	return ret
}

func (mm MembersModel) String() string {
	return strings.Join(mm.Strings(), ", ")
}

// MembersModelAlt is alternative model object representing the current users in the room, but
// with the opposite set of design choices:
// 1) Copy-on-write instead of copy-on-read.
// 2) Applies mutation events rather than direct mutators.
// Contrast with MembersModel.
type MembersModelAlt []string

func (mma MembersModelAlt) ApplyMutation(evt *Event) MembersModelAlt {
	switch evt.What {
	case What_JOIN:
		// Create a new version with a new member.
		ret := append(MembersModelAlt{evt.Who}, mma...)
		sort.Strings(ret)
		return ret
	case What_LEAVE:
		// Create a new version with the member filtered out.
		ret := make(MembersModelAlt, 0, len(mma))
		for _, v := range mma {
			if v != evt.Who {
				ret = append(ret, v)
			}
		}
		return ret
	default:
		return mma // other events are no-ops
	}
}

// Strings returns a copy of the current list of members.
func (mm MembersModelAlt) Strings() []string {
	return mm
}

func (mm MembersModelAlt) String() string {
	return strings.Join(mm, ", ")
}
