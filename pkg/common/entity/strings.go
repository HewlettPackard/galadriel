package entity

import "fmt"

const indent = "    "

func (td *TrustDomain) String() string {
	return fmt.Sprintf(`TrustDomain:
%sID: %s
%sName: %s
%sDescription: %s
%sCreatedAt: %s
%sUpdatedAt: %s`,
		indent, td.ID.UUID,
		indent, td.Name,
		indent, td.Description,
		indent, td.CreatedAt,
		indent, td.UpdatedAt)
}

func (td *TrustDomain) ConsoleString() string {
	return fmt.Sprintf(`TrustDomain:
%sID: %s
%sName: %s
%sDescription: %s`,
		indent, td.ID.UUID,
		indent, td.Name,
		indent, td.Description)
}

func (rel *Relationship) String() string {
	return fmt.Sprintf(`Relationship:
%sID: %s
%sTrustDomainAID: %s
%sTrustDomainBID: %s
%sTrustDomainAName: %s
%sTrustDomainBName: %s
%sTrustDomainAConsent: %s
%sTrustDomainBConsent: %s
%sCreatedAt: %s
%sUpdatedAt: %s`,
		indent, rel.ID.UUID,
		indent, rel.TrustDomainAID,
		indent, rel.TrustDomainBID,
		indent, rel.TrustDomainAName,
		indent, rel.TrustDomainBName,
		indent, rel.TrustDomainAConsent,
		indent, rel.TrustDomainBConsent,
		indent, rel.CreatedAt,
		indent, rel.UpdatedAt)
}

func (rel *Relationship) ConsoleString() string {
	return fmt.Sprintf(`Relationship:
%sID: %s
%sTrust Domain A: %s
%sTrust Domain A Consent Status: %s
%sTrust Domain B: %s
%sTrust Domain B Consent Status: %s`,
		indent, rel.ID.UUID,
		indent, rel.TrustDomainAName,
		indent, rel.TrustDomainAConsent,
		indent, rel.TrustDomainBName,
		indent, rel.TrustDomainBConsent)
}

func (jt *JoinToken) String() string {
	return fmt.Sprintf(`JoinToken:
%sID: %s
%sToken: %s
%sUsed: %t
%sTrustDomainID: %s
%sTrustDomainName: %s
%sExpiresAt: %s
%sCreatedAt: %s
%sUpdatedAt: %s`,
		indent, jt.ID.UUID,
		indent, jt.Token,
		indent, jt.Used,
		indent, jt.TrustDomainID,
		indent, jt.TrustDomainName,
		indent, jt.ExpiresAt,
		indent, jt.CreatedAt,
		indent, jt.UpdatedAt)
}

func (jt *JoinToken) ConsoleString() string {
	return fmt.Sprintf("Token: %s\n", jt.Token)
}

func (b *Bundle) String() string {
	return fmt.Sprintf(`Bundle:
%sID: %s
%sData: %s
%sDigest: %s
%sSignature: %s
%sSigningCertificate: %s
%sTrustDomainID: %s
%sTrustDomainName: %s
%sCreatedAt: %s
%sUpdatedAt: %s`,
		indent, b.ID.UUID,
		indent, b.Data,
		indent, b.Digest,
		indent, b.Signature,
		indent, b.SigningCertificate,
		indent, b.TrustDomainID,
		indent, b.TrustDomainName,
		indent, b.CreatedAt,
		indent, b.UpdatedAt)
}

func (b *Bundle) ConsoleString() string {
	return fmt.Sprintf(`Bundle:
%s%s
%sTrust Domain: %s`,
		indent, b.String(),
		indent, b.TrustDomainName)
}
