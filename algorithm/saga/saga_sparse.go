/* -*- mode: go; -*-
 *
 * Copyright (C) 2019 Philipp Benner
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */
package saga
/* -------------------------------------------------------------------------- */
/* -------------------------------------------------------------------------- */
/* -------------------------------------------------------------------------- */
import "fmt"
import "math"
import "math/rand"
import . "github.com/pbenner/autodiff"
/* -------------------------------------------------------------------------- */
/* -------------------------------------------------------------------------- */
type ObjectiveSparse func(int, *SparseBareRealVector) (ConstReal, *SparseBareRealVector, ConstReal, error)
type InSituSparse struct {
  T1 *SparseBareRealVector
  T2 *BareReal
}
/* -------------------------------------------------------------------------- */
type GradientSparse struct {
  g *SparseBareRealVector
  w ConstReal
}
func (obj GradientSparse) add(v *SparseBareRealVector) {
  for it := v.JOINT_ITERATOR_(obj.g); it.Ok(); it.Next() {
    s_a, s_b := it.GET()
    if s_a == nil {
      s_a = v.AT(it.Index())
    }
    if s_b != nil {
      s_a.SetValue(s_a.GetValue() + obj.w.GetValue()*s_b.GetValue())
    }
  }
}
func (obj GradientSparse) sub(v *SparseBareRealVector) {
  for it := v.JOINT_ITERATOR_(obj.g); it.Ok(); it.Next() {
    s_a, s_b := it.GET()
    if s_a == nil {
      s_a = v.AT(it.Index())
    }
    if s_b != nil {
      s_a.SetValue(s_a.GetValue() - obj.w.GetValue()*s_b.GetValue())
    }
  }
}
func (obj *GradientSparse) set(g *SparseBareRealVector, w ConstReal) {
  if obj.g == nil {
    obj.g = NullSparseBareRealVector(g.Dim())
  }
  obj.g.SET(g)
  obj.w = w
}
/* -------------------------------------------------------------------------- */
func l1regularizationSparse(x *SparseBareRealVector, w *SparseBareRealVector, t *BareReal, lambda float64) {
  for i := 0; i < x.Dim(); i++ {
    if yi := w.ValueAt(i); yi < 0.0 {
      x.AT(i).SetValue(-1.0*math.Max(math.Abs(yi) - lambda, 0.0))
    } else {
      x.AT(i).SetValue( 1.0*math.Max(math.Abs(yi) - lambda, 0.0))
    }
  }
}
func l2regularizationSparse(x *SparseBareRealVector, w *SparseBareRealVector, t *BareReal, lambda float64) {
  t.Vnorm(w)
  t.Div(ConstReal(lambda), t)
  t.Sub(ConstReal(1.0), t)
  t.Max(ConstReal(0.0), t)
  x.VMULS(w, t)
}
/* -------------------------------------------------------------------------- */
func sagaSparse(
  f ObjectiveSparse,
  n int,
  x Vector,
  gamma Gamma,
  epsilon Epsilon,
  maxEpochs MaxEpochs,
  maxIterations MaxIterations,
  l1reg L1Regularization,
  l2reg L2Regularization,
  hook Hook,
  inSitu *InSituSparse) (Vector, error) {
  x1 := AsSparseBareRealVector(x)
  x2 := AsSparseBareRealVector(x)
  // length of gradient
  d := x.Dim()
  // gradient
  var g1 GradientSparse
  var g2 GradientSparse
  // allocate temporary memory
  if inSitu.T1 == nil {
    inSitu.T1 = NullSparseBareRealVector(d)
  }
  if inSitu.T2 == nil {
    inSitu.T2 = NullBareReal()
  }
  // temporary variables
  t1 := inSitu.T1
  t2 := inSitu.T2
  // some constants
  t_n := NewBareReal(float64(n))
  t_g := NewBareReal(gamma.Value)
  // sum of gradients
  s := NullSparseBareRealVector(d)
  dict := make([]GradientSparse, n)
  // initialize s and d
  for i := 0; i < n; i++ {
    if _, g, w, err := f(i, x1); err != nil {
      return nil, err
    } else {
      dict[i].set(g, w)
      dict[i].add(s)
    }
  }
  y := ConstReal(math.NaN())
  for i_ := 0; i_ < maxIterations.Value && i_/n < maxEpochs.Value; i_++ {
    // execute hook if available
    if hook.Value != nil && hook.Value(x1, s, y) {
      break
    }
    j := rand.Intn(n)
    // get old gradient
    g1 = dict[j]
    // evaluate objective function
    if y_, g, w, err := f(j, x1); err != nil {
      return x1, err
    } else {
      y = y_
      g2.set(g, w)
    }
    t1.VDIVS(s , t_n)
    g2.add(t1)
    g1.sub(t1)
    t1.VMULS(t1, t_g)
    switch {
    case l1reg.Value != 0.0:
      t1.VSUBV(x1, t1)
      l1regularizationSparse(x2, t1, t2, gamma.Value*l1reg.Value)
    case l2reg.Value != 0.0:
      t1.VSUBV(x1, t1)
      l2regularizationSparse(x2, t1, t2, gamma.Value*l2reg.Value)
    default:
      x2.VSUBV(x1, t1)
    }
    // evaluate stopping criterion
    max_x := 0.0
    max_delta := 0.0
    for it := x1.JOINT_ITERATOR_(x2); it.Ok(); it.Next() {
      s1, s2 := it.GET()
      v1, v2 := 0.0, 0.0
      if s1 != nil {
        v1 = s1.GetValue()
      }
      if s2 != nil {
        v2 = s2.GetValue()
      }
      if math.IsNaN(v2) {
        return x1, fmt.Errorf("NaN value detected")
      }
      max_x = math.Max(max_x , math.Abs(v2))
      max_delta = math.Max(max_delta, math.Abs(v2 - v1))
    }
    if max_x != 0.0 && max_delta/max_x <= epsilon.Value*gamma.Value ||
      (max_x == 0.0 && max_delta == 0.0) {
      return x2, nil
    }
    x1, x2 = x2, x1
    // update table
    g1.sub(s)
    g2.add(s)
    // update dictionary
    dict[j].set(g2.g, g2.w)
  }
  return x1, nil
}