package match

func EmptyIRInfo() StructuredIRInfo {
	return StructuredIRInfo{
		Nullable:          true,
		FirstLiteralExact: true,
		LastLiteralExact:  true,
	}
}

func CombineIR(parts ...StructuredIRInfo) StructuredIRInfo {
	info := EmptyIRInfo()
	if len(parts) == 0 {
		return info
	}
	for _, part := range parts {
		info.MinWidth += part.MinWidth
		if !part.Nullable {
			info.Nullable = false
		}
	}
	info.FirstLiteral, info.FirstLiteralExact = IRBoundaryLiteral(parts, true)
	info.LastLiteral, info.LastLiteralExact = IRBoundaryLiteral(parts, false)
	return info
}

func IRBoundaryLiteral(parts []StructuredIRInfo, forward bool) (string, bool) {
	var candidate string
	haveCandidate := false

	if forward {
		for i := 0; i < len(parts); i++ {
			part := parts[i]
			if part.FirstLiteral != "" && part.FirstLiteralExact {
				if !haveCandidate {
					candidate = part.FirstLiteral
					haveCandidate = true
				} else if candidate != part.FirstLiteral {
					return "", false
				}
			} else if part.Nullable {
				return "", false
			}
			if !part.Nullable {
				if haveCandidate {
					return candidate, true
				}
				return "", false
			}
		}
		return "", false
	}

	for i := len(parts) - 1; i >= 0; i-- {
		part := parts[i]
		if part.LastLiteral != "" && part.LastLiteralExact {
			if !haveCandidate {
				candidate = part.LastLiteral
				haveCandidate = true
			} else if candidate != part.LastLiteral {
				return "", false
			}
		} else if part.Nullable {
			return "", false
		}
		if !part.Nullable {
			if haveCandidate {
				return candidate, true
			}
			return "", false
		}
	}

	return "", false
}
