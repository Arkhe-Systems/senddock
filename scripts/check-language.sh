#!/usr/bin/env bash
# Verify that everything tracked by git is written in English.
#
# English is the language the repository ships in: code, comments, docs, changelog,
# roadmap, examples and sample data. Spanish is fine for working notes that stay out of
# git, but a tracked file that mixes the two is unreadable in review and cannot be
# translated later without guessing which half is current. This is the check that keeps
# that from creeping back in, because it is invisible to a reader who knows both
# languages and invisible to a reader who knows neither.
#
# Translations, if the project ever ships them, get their own files and their own
# review — they are not a reason to relax this. Neither is an accented proper noun in an
# example: sample data uses English names like the rest of the repository, so the fix is
# to use one, not to add an exception here.
#
# Usage:
#   scripts/check-language.sh
set -euo pipefail

root="$(cd "$(dirname "$0")/.." && pwd)"
cd "$root"

# Spanish words that are not English words and are not common identifiers. Left out on
# purpose: tokens English also uses (sin, solo, hay), language codes (es) and code
# identifiers that collide (del). A check that cries wolf gets switched off.
spanish_words='que|para|los|las|una|desde|cuando|donde|porque|tambien|también|siempre|nunca|usuario|usuarios|correo|correos|contraseña|envio|envío|envios|envíos|plantilla|plantillas|pagina|página|paginas|páginas|segun|según|ademas|además|tiene|tienen|puede|pueden|debe|deben|hacer|falta|agregar|arreglar|cambiar|revisar|necesita|entonces|mientras|aunque|deberia|debería|seria|sería|está|están|debido|respecto'

# A word, not a character: this way the report names what to fix, and a single accented
# character inside an otherwise English file is still caught.
accented_word='[A-Za-zÁÉÍÓÚÑáéíóúñ]*[áéíóúüñÁÉÍÓÚÜÑ¿¡][A-Za-zÁÉÍÓÚÑáéíóúñ]*'

# One hit is not a language; three Spanish words in one file is prose. The threshold
# exists so a stray match cannot block a build.
word_threshold=3

failed=0
checked=0

while IFS= read -r file; do
  [ -f "$file" ] || continue
  checked=$((checked + 1))

  # -I keeps binaries out: a screenshot is not prose and cannot be translated.
  words="$(grep -EIow "$spanish_words" -- "$file" 2>/dev/null | sort | uniq -c | sort -rn || true)"
  word_count="$(printf '%s\n' "$words" | awk '{total += $1} END {print total + 0}')"
  accents="$(grep -EIo "$accented_word" -- "$file" 2>/dev/null | sort -u | paste -sd' ' - || true)"

  if [ "$word_count" -lt "$word_threshold" ] && [ -z "$accents" ]; then
    continue
  fi

  failed=$((failed + 1))
  {
    printf '  %s\n' "$file"
    [ -n "$accents" ] && printf '    accented words: %s\n' "$accents"
    if [ "$word_count" -ge "$word_threshold" ]; then
      printf '    spanish words:  %s\n' "$(printf '%s\n' "$words" | awk '{printf "%s(x%s) ", $2, $1}')"
    fi
  } >&2
done < <(git ls-files)

if [ "$failed" -gt 0 ]; then
  printf '\ncheck-language: %s of %s tracked files are not in English (listed above).\n' "$failed" "$checked" >&2
  printf 'Rewrite them in English. Internal notes in Spanish do not belong in the repository.\n' >&2
  exit 1
fi

printf 'check-language: OK — %s tracked files, all in English\n' "$checked"
