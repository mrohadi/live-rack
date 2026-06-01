// Zitadel project-roles claim: { roleName: { orgId: orgDomain } }.
export const ROLES_CLAIM = "urn:zitadel:iam:org:project:roles";

/** Role names granted to the user, from the OIDC project-roles claim. Pure. */
export function rolesFromProfile(profile?: Record<string, unknown>): string[] {
  const roles = profile?.[ROLES_CLAIM] as Record<string, unknown> | undefined;
  return roles ? Object.keys(roles) : [];
}

/** True when the profile carries the admin role. Pure. */
export function isAdmin(profile?: Record<string, unknown>): boolean {
  return rolesFromProfile(profile).includes("admin");
}
