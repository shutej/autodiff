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

import   "fmt"

import . "github.com/pbenner/autodiff/statistics"
import   "github.com/pbenner/autodiff/statistics/generic"

import . "github.com/pbenner/autodiff"

/* -------------------------------------------------------------------------- */

type Chmm struct {
  Hmm
}

/* -------------------------------------------------------------------------- */

func NewConstrainedHmm(pi Vector, tr Matrix, stateMap []int, edist []ScalarPdf, tree generic.HmmNode) (*Chmm, error) {
  return newConstrainedHmm(pi, tr, stateMap, edist, tree, false)
}

func newConstrainedHmm(pi Vector, tr Matrix, stateMap []int, edist []ScalarPdf, tree generic.HmmNode, isLog bool) (*Chmm, error) {
  if !tree.Check(pi.Dim()) {
    return nil, fmt.Errorf("invalid Hmm tree")
  }
  p, err := generic.NewHmmProbabilityVector(pi, isLog); if err != nil {
    return nil, err
  }
  t, err := generic.NewChmmTransitionMatrix(tr, tree, isLog); if err != nil {
    return nil, err
  }
  if hmm, err := generic.NewHmm(p, t, stateMap); err != nil {
    return nil, err
  } else {
    if len(edist) == 0 {
      edist = make([]ScalarPdf, hmm.NEDists())
    } else {
      if hmm.NEDists() != len(edist) {
        return nil, fmt.Errorf("invalid number of emission distributions")
      }
    }
    return &Chmm{Hmm{*hmm, edist}}, nil
  }
}

/* -------------------------------------------------------------------------- */

func (obj *Chmm) ImportConfig(config ConfigDistribution, t ScalarType) error {
  hmm  := Hmm{}
  tree := generic.HmmNode{}
  if err := hmm.ImportConfig(config, t); err != nil {
    return err
  }
  p, ok := config.GetNamedParameter("Tree"); if !ok {
    return fmt.Errorf("invalid config file")
  }
  if ok := tree.ImportConfig(p); !ok {
    return fmt.Errorf("invalid config file")
  }
  // get hierarchical transition matrix
  if r, err := newConstrainedHmm(hmm.Pi, hmm.Tr, hmm.StateMap, hmm.Edist, tree, true); err != nil {
    return err
  } else {
    *obj = *r
  }
  return nil
}

func (obj *Chmm) ExportConfig() ConfigDistribution {
  configHmm := obj.Hmm.ExportConfig()
  parametersHmm := configHmm.Parameters.(struct{
    Pi          []float64
    Tr          []float64
    StateMap    []int
    N             int
    StartStates []int
    FinalStates []int })

  parameters := struct{
    Pi          []float64
    Tr          []float64
    StateMap    []int
    N             int
    StartStates []int
    FinalStates []int
    Tree          interface{} }{}

  parameters.Pi          = parametersHmm.Pi
  parameters.Tr          = parametersHmm.Tr
  parameters.StateMap    = parametersHmm.StateMap
  parameters.N           = parametersHmm.N
  parameters.StartStates = parametersHmm.StartStates
  parameters.FinalStates = parametersHmm.FinalStates
  parameters.Tree        = obj.Tr.(generic.ChmmTransitionMatrix).Tree.ExportConfig()

  return NewConfigDistribution("vector:constrained hmm distribution", parameters, configHmm.Distributions...)
}