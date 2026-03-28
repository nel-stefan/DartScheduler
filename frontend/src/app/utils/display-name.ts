/** Converts "Achternaam, Voornaam" to "Voornaam Achternaam" for display. Mirrors backend domain.FormatDisplayName. */
export function displayName(name: string): string {
  const idx = name.indexOf(', ');
  return idx >= 0 ? `${name.slice(idx + 2).trim()} ${name.slice(0, idx).trim()}` : name;
}
