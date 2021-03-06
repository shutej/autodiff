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

package vectorDistribution

/* -------------------------------------------------------------------------- */

//import   "fmt"

import . "github.com/pbenner/autodiff/statistics"

import . "github.com/pbenner/autodiff"

/* -------------------------------------------------------------------------- */

type HmmDataRecord struct {
  Edist []ScalarPdf
  X       ConstVector
}

func (obj HmmDataRecord) MapIndex(k int) int {
  return k
}

func (obj HmmDataRecord) GetN() int {
  return obj.X.Dim()
}

func (obj HmmDataRecord) LogPdf(r Scalar, c, k int) error {
  return obj.Edist[c].LogPdf(r, obj.X.ConstAt(k))
}
