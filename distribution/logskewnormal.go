/* Copyright (C) 2016 Philipp Benner
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

package distribution

/* -------------------------------------------------------------------------- */

//import   "fmt"
import   "math"

import . "github.com/pbenner/autodiff"

/* -------------------------------------------------------------------------- */

type LogSkewNormalDistribution struct {
  Normal1 NormalDistribution
  Normal2 NormalDistribution
  Omega   Matrix
  Alpha   Vector
  Scale   Vector
}

/* -------------------------------------------------------------------------- */

func NewLogSkewNormalDistribution(xi Vector, omega Matrix, alpha, scale Vector) (*LogSkewNormalDistribution, error) {
  // dimension
  n, m := omega.Dims()
  // check parameter dimensions
  if n != len(xi)     ||
    (n != len(alpha)) ||
    (n != len(scale)) ||
    (n != m) {
    panic("NewLogSkewNormalDistribution(): Parameter dimensions do not match!")
  }
  // parameters for the multivariate normal
  // kappa = diag(s) omega diag(s)
  kappa := NullMatrix(RealType, n, n)
  for i := 0; i < n; i++ {
    for j := 0; j < n; j++ {
      kappa.At(i, j).Mul(Mul(scale[i], scale[j]), omega.At(i,j))
    }
  }
  // parameters for the standard normal cdf
  mu    := NewVector(RealType,       []float64{0})
  sigma := NewMatrix(RealType, 1, 1, []float64{1})

  normal1, err := NewNormalDistribution(xi, kappa)
  if err != nil { return nil, err }
  normal2, err := NewNormalDistribution(mu, sigma)
  if err != nil { return nil, err }

  result := LogSkewNormalDistribution{*normal1, *normal2, omega, alpha, scale}

  return &result, nil

}

/* -------------------------------------------------------------------------- */


func (dist *LogSkewNormalDistribution) Clone() *LogSkewNormalDistribution {
  return &LogSkewNormalDistribution{
    Normal1: *dist.Normal1.Clone(),
    Normal2: *dist.Normal2.Clone(),
    Omega  :  dist.Omega  .Clone(),
    Alpha  :  dist.Alpha  .Clone(),
    Scale  :  dist.Scale  .Clone() }
}

func (dist LogSkewNormalDistribution) Dim() int {
  return len(dist.Alpha)
}

func (dist LogSkewNormalDistribution) LogPdf(x Vector) Scalar {
  n := dist.Normal1.Dim()
  c := NewScalar(RealType, math.Log(2))
  y := NullVector(RealType, n)
  z := NullVector(RealType, n)
  t := NullVector(RealType, 1)
  for i := 0; i < n; i++ {
    y[i] = Log(Add(x[i], NewReal(1.0)))
    z[i] = Div(Sub(y[i], dist.Normal1.Mu[i]), dist.Scale[i])
  }
  t[0] = VdotV(dist.Alpha, z)
  // add the det Jacobian of the variable transform to the constant c
  for i := 0; i < n; i++ {
    c = Add(c, Neg(Log(y[i])))
  }

  return Add(c,
    Add(dist.Normal1.LogPdf(y), dist.Normal2.LogCdf(t)))
}

func (dist LogSkewNormalDistribution) Pdf(x Vector) Scalar {
  return Exp(dist.LogPdf(x))
}

func (dist LogSkewNormalDistribution) Cdf(x Vector) Scalar {
  panic("Method not implemented!")
}

func (dist LogSkewNormalDistribution) LogCdf(x Vector) Scalar {
  panic("Method not implemented!")
}

/* -------------------------------------------------------------------------- */

func (dist *LogSkewNormalDistribution) GetParameters() Vector {
  p := dist.Normal1.Mu
  p  = append(p, dist.Omega.Vector()...)
  p  = append(p, dist.Alpha...)
  p  = append(p, dist.Scale...)
  return p
}

func (dist *LogSkewNormalDistribution) SetParameters(parameters Vector) error {
  n := dist.Dim()
  xi    := parameters[0*n+0*n*n:1*n+0*n*n]
  omega := parameters[1*n+0*n*n:1*n+1*n*n].Matrix(n, n)
  alpha := parameters[1*n+1*n*n:2*n+1*n*n]
  scale := parameters[2*n+1*n*n:3*n+1*n*n]
  if tmp, err := NewLogSkewNormalDistribution(xi, omega, alpha, scale); err != nil {
    return err
  } else {
    *dist = *tmp
  }
  return nil
}
