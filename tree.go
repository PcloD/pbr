package pbr

import (
	"fmt"
	"math"
	"sort"
	"strconv"
)

// https://people.cs.clemson.edu/~dhouse/courses/405/notes/KDtrees-Fussell.pdf
// http://slideplayer.com/slide/7653218/
type Tree struct {
	box      *Box
	left     *Tree
	right    *Tree
	surfaces []Surface
	leaf     bool
	name     string
}

func NewTree(box *Box, surfaces []Surface, depth int, suffix string) *Tree {
	t := &Tree{
		surfaces: overlaps(box, surfaces),
		box:      box,
		name:     strconv.Itoa(depth) + suffix,
	}
	t.name += " (" + fmt.Sprintf("%v", &t) + ")"
	limit := int(math.Pow(2, float64(depth-1)))
	if len(t.surfaces) < limit || depth > 10 {
		t.leaf = true
		return t
	}
	axis := depth % 3
	wall := median(t.surfaces, axis)
	lBox, rBox := box.Split(axis, wall)
	t.left = NewTree(lBox, t.surfaces, depth+1, "L")
	t.right = NewTree(rBox, t.surfaces, depth+1, "R")
	return t
}

// TODO: implement as Box.Overlaps(Box) and replace Surface.Bounds() with Surface.Box() which actually returns a Box
func overlaps(box *Box, surfaces []Surface) []Surface {
	matches := make([]Surface, 0)
	for _, s := range surfaces {
		if s.Box().Overlaps(box) {
			matches = append(matches, s)
		}
	}
	return matches
}

// http://slideplayer.com/slide/7653218/
func (t *Tree) Intersect(ray *Ray3) Hit {
	if t.leaf {
		hit := t.IntersectSurfaces(ray)
		return hit
	}
	left, lDist := t.left.Check(ray)
	right, rDist := t.right.Check(ray)
	if left && right {
		if lDist < rDist {
			hit := t.left.Intersect(ray) // TODO: can probably optimize Intersect by putting a limit on the ray length, then skip this dist test
			if hit.ok {                  // TODO: instead of checking for rDist, get tmin and tmax from .Check() and check against tmax
				return hit
			}
			return t.right.Intersect(ray)
		}
		hit := t.right.Intersect(ray)
		if hit.ok {
			return hit
		}
		return t.left.Intersect(ray)
	} else if left {
		return t.left.Intersect(ray)
	} else if right {
		return t.right.Intersect(ray)
	}
	return Miss
}

func (t *Tree) IntersectSurfaces(ray *Ray3) Hit {
	closest := Miss
	for _, s := range t.surfaces {
		hit := s.Intersect(ray)
		if hit.ok {
			if t.box.Contains(ray.Moved(hit.dist)) {
				closest = hit.Closer(closest)
			} else {
			}
		}
	}
	return closest
}

func (t *Tree) Check(ray *Ray3) (ok bool, dist float64) {
	return t.box.Check(ray)
}

func median(surfaces []Surface, axis int) float64 {
	vals := make([]float64, 0)
	for _, s := range surfaces {
		b := s.Box()
		vals = append(vals, b.minArray[axis], b.maxArray[axis])
	}
	sort.Float64s(vals)
	return vals[len(vals)/2]
}
