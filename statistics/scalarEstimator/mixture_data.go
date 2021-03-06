/* Copyright (C) 2017 Philipp Benner
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

package scalarEstimator

/* -------------------------------------------------------------------------- */

import   "fmt"
import   "math"

import . "github.com/pbenner/autodiff/statistics"
import   "github.com/pbenner/autodiff/statistics/generic"
import . "github.com/pbenner/autodiff/logarithmetic"

import . "github.com/pbenner/autodiff"
import . "github.com/pbenner/threadpool"

/* -------------------------------------------------------------------------- */

type MixtureDataSet interface {
  generic.MixtureDataSet
  GetData () ConstVector
  EvaluateLogPdf(edist []ScalarPdf, pool ThreadPool) error
}

/* -------------------------------------------------------------------------- */

type MixtureStdDataSet struct {
  values ConstVector
  n      int
  p      Matrix
}

func NewMixtureStdDataSet(t ScalarType, x ConstVector, k int) (*MixtureStdDataSet, error) {
  r := MixtureStdDataSet{}
  r.values = x
  r.p      = NullMatrix(t, k, x.Dim())
  r.n      = x.Dim()
  return &r, nil
}

func (obj *MixtureStdDataSet) GetData() ConstVector {
  return obj.values
}

func (obj *MixtureStdDataSet) GetCounts() []int {
  return nil
}

func (obj *MixtureStdDataSet) GetN() int {
  return obj.n
}

func (obj *MixtureStdDataSet) LogPdf(r Scalar, c, i int) error {
  r.Set(obj.p.At(c, i))
  return nil
}

func (obj *MixtureStdDataSet) EvaluateLogPdf(edist []ScalarPdf, pool ThreadPool) error {
  x    := obj.values
  p    := obj.p
  m, n := obj.p.Dims()
  if len(edist) != m {
    return fmt.Errorf("data has invalid dimension")
  }
  // distributions may have state and must be cloned
  // for each thread
  d := make([][]ScalarPdf, pool.NumberOfThreads())
  s := make([]float64, pool.NumberOfThreads())
  for threadIdx := 0; threadIdx < pool.NumberOfThreads(); threadIdx++ {
    d[threadIdx] = make([]ScalarPdf, m)
    for j := 0; j < m; j++ {
      d[threadIdx][j] = edist[j].CloneScalarPdf()
    }
  }
  g := pool.NewJobGroup()
  // evaluate emission distributions
  if err := pool.AddRangeJob(0, n, g, func(i int, pool ThreadPool, erf func() error) error {
    if erf() != nil {
      return nil
    }
    s := s[pool.GetThreadId()]
    d := d[pool.GetThreadId()]
    s = math.Inf(-1)
    // loop over emission distributions
    for j := 0; j < m; j++ {
      if err := d[j].LogPdf(p.At(j, i), x.ConstAt(i)); err != nil {
        return err
      }
      s = LogAdd(s, p.At(j, i).GetValue())
    }
    if math.IsInf(s, -1) {
      return fmt.Errorf("probability is zero for all models on observation `%v'", x.ConstAt(i))
    }
    return nil
  }); err != nil {
    return fmt.Errorf("evaluating emission probabilities failed: %v", err)
  }
  if err := pool.Wait(g); err != nil {
    return fmt.Errorf("evaluating emission probabilities failed: %v", err)
  }
  return nil
}

/* -------------------------------------------------------------------------- */

type MixtureSummarizedDataSet struct {
  values []float64
  counts []int
  p        Matrix
}

func NewMixtureSummarizedDataSet(t ScalarType, x ConstVector, k int) (*MixtureSummarizedDataSet, error) {
  xMap   := make(map[float64]int)
  values := []float64{}
  counts := []int{}
  for i := 0; i < x.Dim(); i++ {
    v := x.ConstAt(i).GetValue()
    if idx, ok := xMap[v]; ok {
      counts[idx] += 1
    } else {
      idx     := len(values)
      values   = append(values, v)
      counts   = append(counts, 1)
      xMap [v] = idx
    }
  }
  r := MixtureSummarizedDataSet{}
  r.values = values
  r.counts = counts
  r.p      = NullMatrix(t, k, len(values))
  return &r, nil
}

func (obj *MixtureSummarizedDataSet) GetData() ConstVector {
  return DenseConstRealVector(obj.values)
}

func (obj *MixtureSummarizedDataSet) GetCounts() []int {
  return obj.counts
}

func (obj *MixtureSummarizedDataSet) GetN() int {
  return len(obj.values)
}

func (obj *MixtureSummarizedDataSet) LogPdf(r Scalar, c, i int) error {
  r.Set(obj.p.At(c, i))
  return nil
}

func (obj *MixtureSummarizedDataSet) EvaluateLogPdf(edist []ScalarPdf, pool ThreadPool) error {
  x    := obj.values
  p    := obj.p
  m, n := obj.p.Dims()
  if len(edist) != m {
    return fmt.Errorf("data has invalid dimension")
  }
  // distributions may have state and must be cloned
  // for each thread
  d := make([][]ScalarPdf, pool.NumberOfThreads())
  s := make([]float64, pool.NumberOfThreads())
  for threadIdx := 0; threadIdx < pool.NumberOfThreads(); threadIdx++ {
    d[threadIdx] = make([]ScalarPdf, m)
    for j := 0; j < m; j++ {
      d[threadIdx][j] = edist[j].CloneScalarPdf()
    }
  }
  g := pool.NewJobGroup()
  // evaluate emission distributions
  if err := pool.AddRangeJob(0, n, g, func(i int, pool ThreadPool, erf func() error) error {
    if erf() != nil {
      return nil
    }
    s := s[pool.GetThreadId()]
    d := d[pool.GetThreadId()]
    s = math.Inf(-1)
    // loop over emission distributions
    for j := 0; j < m; j++ {
      if err := d[j].LogPdf(p.At(j, i), ConstReal(x[i])); err != nil {
        return err
      }
      s = LogAdd(s, p.At(j, i).GetValue())
    }
    if math.IsInf(s, -1) {
      return fmt.Errorf("probability is zero for all models on observation `%v'", x[i])
    }
    return nil
  }); err != nil {
    return fmt.Errorf("evaluating emission probabilities failed: %v", err)
  }
  if err := pool.Wait(g); err != nil {
    return fmt.Errorf("evaluating emission probabilities failed: %v", err)
  }
  return nil
}
