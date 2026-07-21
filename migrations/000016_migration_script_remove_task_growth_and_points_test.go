package migrations

import (
	"reflect"
	"testing"
)

func TestRemoveTaskNavigationItems(t *testing.T) {
	items := []interface{}{
		map[string]interface{}{"title": "Topics", "url": "/topics"},
		map[string]interface{}{"title": "Tasks", "url": "/tasks"},
		map[string]interface{}{
			"title": "More",
			"children": []interface{}{
				map[string]interface{}{"title": "Tasks", "url": "/tasks"},
				map[string]interface{}{"title": "Articles", "url": "/articles"},
			},
		},
		map[string]interface{}{
			"title": "Retired",
			"children": []interface{}{
				map[string]interface{}{"title": "Tasks", "url": "/tasks"},
			},
		},
	}

	got := removeTaskNavigationItems(items)
	want := []interface{}{
		map[string]interface{}{"title": "Topics", "url": "/topics"},
		map[string]interface{}{
			"title": "More",
			"children": []interface{}{
				map[string]interface{}{"title": "Articles", "url": "/articles"},
			},
		},
		map[string]interface{}{"title": "Retired"},
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected filtered navigation: %#v", got)
	}
}
