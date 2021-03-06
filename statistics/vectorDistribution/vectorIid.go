/* Copyright (C) 2016-2017 Philipp Benner
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

package vectorDistribution

/* -------------------------------------------------------------------------- */

import   "fmt"

import . "github.com/pbenner/autodiff/statistics"

import . "github.com/pbenner/autodiff"

/* -------------------------------------------------------------------------- */

type VectorIid struct {
  Distribution VectorPdf
  n int
  t Scalar
}

/* -------------------------------------------------------------------------- */

func NewVectorIid(distribution VectorPdf, n int) (*VectorIid, error) {
  m := distribution.Dim()
  t := NewScalar(distribution.ScalarType(), 0.0)
  if n < 0 || n % m != 0 {
    return nil, fmt.Errorf("error while creating a vector iid distribution: dimension `%d' is not a multiple of dimension `%d'", n, m)
  }
  return &VectorIid{distribution, n, t}, nil
}

/* -------------------------------------------------------------------------- */

func (obj *VectorIid) Clone() *VectorIid {
  return &VectorIid{obj.Distribution.CloneVectorPdf(), obj.n, obj.t.CloneScalar()}
}

func (obj *VectorIid) CloneVectorPdf() VectorPdf {
  return obj.Clone()
}

/* -------------------------------------------------------------------------- */

func (obj *VectorIid) Dim() int {
  return obj.n
}

func (obj *VectorIid) ScalarType() ScalarType {
  return obj.Distribution.ScalarType()
}

func (obj *VectorIid) LogPdf(r Scalar, x ConstVector) error {
  n := obj.Dim()
  m := obj.Distribution.Dim()
  t := obj.t
  if x.Dim() != n {
    return fmt.Errorf("LogPdf(): dimensions do not match (input has dimension `%d' whereas this distribution is of dimension `%d'", x.Dim(), obj.Dim())
  }
  r.Reset()
  for i := 0; i < n; i += m {
    if err := obj.Distribution.LogPdf(t, x.ConstSlice(i,i+m)); err != nil {
      return err
    }
    r.Add(r, t)
  }
  return nil
}

/* -------------------------------------------------------------------------- */

func (obj *VectorIid) GetParameters() Vector {
  return obj.Distribution.GetParameters()
}

func (obj *VectorIid) SetParameters(parameters Vector) error {
  return obj.Distribution.SetParameters(parameters)
}

/* -------------------------------------------------------------------------- */

func (obj *VectorIid) ImportConfig(config ConfigDistribution, t ScalarType) error {

  if parameters, ok := config.GetParametersAsFloats(); !ok {
    return fmt.Errorf("invalid config file")
  } else {
    if len(parameters) != 1 {
      return fmt.Errorf("invalid config file")
    }
    if len(config.Distributions) != 1 {
      return fmt.Errorf("invalid config file")
    }

    if dist, err := ImportVectorPdfConfig(config.Distributions[0], t); err != nil {
      return err
    } else {
      if tmp, err := NewVectorIid(dist, int(parameters[0])); err != nil {
        return err
      } else {
        *obj = *tmp
      }
    }
    return nil
  }
}

func (obj *VectorIid) ExportConfig() (config ConfigDistribution) {

  parameters := NewVector(BareRealType, []float64{float64(obj.n)})

  return NewConfigDistribution("vector:vector iid", parameters, obj.Distribution.ExportConfig())
}
