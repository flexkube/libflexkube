package flexkube

import (
	"sort"

	"github.com/flexkube/libflexkube/pkg/container"
)

func containersStateUnmarshal(i interface{}) *container.ContainersState {
	hccs := i.([]interface{})

	if len(hccs) == 0 {
		return nil
	}

	cs := container.ContainersState{}

	for _, hcc := range hccs {
		n, h := hostConfiguredContainerUnmarshal(hcc)
		cs[n] = h
	}

	return &cs
}

func containersStateMarshal(c container.ContainersState, sensitive bool) []interface{} {
	names := []string{}

	for i := range c {
		names = append(names, i)
	}

	sort.Strings(names)

	var r []interface{} //nolint:prealloc

	for _, n := range names {
		r = append(r, hostConfiguredContainerMarshal(n, *c[n], sensitive))
	}

	return r
}
