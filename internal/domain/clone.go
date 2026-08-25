package domain

func (t *Task) Clone() *Task {
	out := *t
	out.Configs = map[string]*Config{}
	for k, v := range t.Configs {
		out.Configs[k] = v.Clone()
	}
	out.Risks = map[string]*Risk{}
	for k, v := range t.Risks {
		x := *v
		out.Risks[k] = &x
	}
	if t.Permit != nil {
		x := *t.Permit
		out.Permit = &x
	}
	if t.PermitHistory != nil {
		out.PermitHistory = make([]*Permit, 0, len(t.PermitHistory))
		for _, p := range t.PermitHistory {
			if p == nil {
				continue
			}
			x := *p
			if p.RevokedAt != nil {
				v := *p.RevokedAt
				x.RevokedAt = &v
			}
			out.PermitHistory = append(out.PermitHistory, &x)
		}
	}
	if t.ValidatedReports != nil {
		out.ValidatedReports = map[string]ValidationRecord{}
		for k, v := range t.ValidatedReports {
			v.Findings = append([]Finding(nil), v.Findings...)
			out.ValidatedReports[k] = v
		}
	}
	return &out
}

func (r *Risk) Clone() *Risk {
	if r == nil {
		return nil
	}
	x := *r
	return &x
}
