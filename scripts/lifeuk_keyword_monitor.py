#!/usr/bin/env python3

import json
import os
import re
import subprocess
import sys
from datetime import date, datetime, timezone
from pathlib import Path


REPO_ROOT = Path(__file__).resolve().parents[1]
SEARCHADS_BIN = REPO_ROOT / "bin" / "searchads"
OUTPUT_ROOT = REPO_ROOT / "reports" / "lifeuk-monitor"
MONITOR_START_DATE = "2026-03-07"

DISCOVERY_CAMPAIGN_ID = 2143463411
DISCOVERY_BROAD_ADGROUP_ID = 2146343714
DISCOVERY_SEARCHMATCH_ADGROUP_ID = 2146344840
CATEGORY_CAMPAIGN_ID = 2143492917
CATEGORY_EXACT_ADGROUP_ID = 2146757366
PROMOTION_BID_GBP = "0.75"

HIGH_INTENT_MARKERS = (
    "test",
    "practice",
    "prep",
    "exam",
    "mock",
    "revision",
    "citizenship",
    "life in the uk",
)


def ensure_credentials(env: dict[str, str]) -> dict[str, str]:
    if env.get("OE_ADS_CREDENTIALS_JSON"):
        return env

    creds_file = env.get("OE_ADS_CREDENTIALS_FILE", "").strip()
    if creds_file:
        path = Path(creds_file).expanduser()
        if path.exists():
            env["OE_ADS_CREDENTIALS_JSON"] = path.read_text(encoding="utf-8").strip()
            return env

    result = subprocess.run(
        [
            "security",
            "find-generic-password",
            "-s",
            "OpportunityEngine.AppleAdsCredentials",
            "-a",
            "default",
            "-w",
        ],
        capture_output=True,
        text=True,
        check=True,
    )
    env["OE_ADS_CREDENTIALS_JSON"] = result.stdout.strip()
    return env


def run_searchads(*args: str) -> dict:
    env = ensure_credentials(os.environ.copy())
    cmd = [str(SEARCHADS_BIN), *args, "--json"]
    result = subprocess.run(
        cmd,
        cwd=REPO_ROOT,
        env=env,
        capture_output=True,
        text=True,
    )
    if result.returncode != 0:
        raise RuntimeError(
            f"Command failed ({result.returncode}): {' '.join(cmd)}\n"
            f"stdout={result.stdout}\nstderr={result.stderr}"
        )
    return json.loads(result.stdout)


def normalize_term(text: str) -> str:
    return re.sub(r"\s+", " ", text.strip().lower())


def token_count(text: str) -> int:
    return len(re.findall(r"[a-z0-9]+", text.lower()))


def high_intent(text: str) -> bool:
    normalized = normalize_term(text)
    if any(marker in normalized for marker in HIGH_INTENT_MARKERS):
        return True
    has_uk = "uk" in normalized
    has_citizenship = "citizenship" in normalized or "british" in normalized
    return has_uk and has_citizenship


def enrich_keyword_report(payload: dict) -> dict:
    rows = []
    for row in payload.get("rows", []):
        item = dict(row)
        item["cr"] = item.get("installRate", 0)
        rows.append(item)

    totals = dict(payload.get("totals", {}))
    totals["cr"] = totals.get("installRate", 0)

    return {
        "campaignId": payload.get("campaignId"),
        "adGroupId": payload.get("adGroupId"),
        "startDate": payload.get("startDate"),
        "endDate": payload.get("endDate"),
        "totals": totals,
        "rows": rows,
    }


def promotion_candidates(
    search_terms: dict,
    existing_exact_terms: set[str],
    existing_negative_terms: set[str],
) -> tuple[list[dict], list[dict]]:
    candidates: list[dict] = []
    skipped: list[dict] = []

    for row in search_terms.get("rows", []):
        term = row.get("searchTerm", "").strip()
        normalized = normalize_term(term)
        reasons: list[str] = []
        taps = int(row.get("taps", 0) or 0)
        installs = int(row.get("installs", 0) or 0)
        spend = float(row.get("spend", 0) or 0)
        cpt = float(row.get("cpt", 0) or 0)
        ttr = float(row.get("ttr", 0) or 0)
        cr = float(row.get("installRate", 0) or 0)

        if not term:
            continue
        if normalized in existing_exact_terms:
            reasons.append("already exact")
        if token_count(term) < 3:
            reasons.append("not long tail")
        if not high_intent(term):
            reasons.append("not clearly high intent")
        if taps < 1:
            reasons.append("no taps")
        if spend > 2.25:
            reasons.append("spend above 2.25")
        if cpt > 0.75:
            reasons.append("cpt above 0.75")
        if installs < 1 and not (taps >= 2 and cr >= 0.25):
            reasons.append("insufficient conversion evidence")
        if ttr < 0.04:
            reasons.append("ttr below 4%")

        if reasons:
            skipped.append(
                {
                    "searchTerm": term,
                    "taps": taps,
                    "installs": installs,
                    "spend": spend,
                    "cpt": cpt,
                    "ttr": ttr,
                    "cr": cr,
                    "reasons": reasons,
                }
            )
            continue

        candidates.append(
            {
                "searchTerm": term,
                "normalized": normalized,
                "taps": taps,
                "installs": installs,
                "spend": spend,
                "cpt": cpt,
                "ttr": ttr,
                "cr": cr,
                "alreadyNegative": normalized in existing_negative_terms,
            }
        )

    candidates.sort(key=lambda row: (-row["installs"], row["cpt"], row["spend"], -row["ttr"]))
    skipped.sort(key=lambda row: (-row["installs"], row["cpt"], row["spend"], row["searchTerm"]))
    return candidates, skipped


def write_json(path: Path, payload: dict) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text(json.dumps(payload, indent=2, sort_keys=True) + "\n")


def main() -> int:
    if not SEARCHADS_BIN.exists():
        print(f"searchads binary not found at {SEARCHADS_BIN}", file=sys.stderr)
        return 1

    checkpoint = datetime.now(timezone.utc)
    checkpoint_slug = checkpoint.strftime("%Y%m%dT%H%M%SZ")
    today = date.today().isoformat()
    run_dir = OUTPUT_ROOT / "checkpoints" / checkpoint_slug
    run_dir.mkdir(parents=True, exist_ok=True)

    broad_report = enrich_keyword_report(
        run_searchads(
            "keywords",
            "report",
            "--campaignId",
            str(DISCOVERY_CAMPAIGN_ID),
            "--adGroupId",
            str(DISCOVERY_BROAD_ADGROUP_ID),
            "--startDate",
            MONITOR_START_DATE,
            "--endDate",
            today,
        )
    )
    exact_report = enrich_keyword_report(
        run_searchads(
            "keywords",
            "report",
            "--campaignId",
            str(CATEGORY_CAMPAIGN_ID),
            "--adGroupId",
            str(CATEGORY_EXACT_ADGROUP_ID),
            "--startDate",
            MONITOR_START_DATE,
            "--endDate",
            today,
        )
    )
    search_match_report = run_searchads(
        "searchterms",
        "report",
        "--campaignId",
        str(DISCOVERY_CAMPAIGN_ID),
        "--adGroupId",
        str(DISCOVERY_SEARCHMATCH_ADGROUP_ID),
        "--startDate",
        MONITOR_START_DATE,
        "--endDate",
        today,
    )

    exact_keywords = run_searchads(
        "keywords",
        "list",
        "--campaignId",
        str(CATEGORY_CAMPAIGN_ID),
        "--adGroupId",
        str(CATEGORY_EXACT_ADGROUP_ID),
    )
    negatives = run_searchads("negatives", "list", "--campaignId", str(DISCOVERY_CAMPAIGN_ID))

    existing_exact_terms = {normalize_term(item["text"]) for item in exact_keywords}
    existing_negative_terms = {normalize_term(item["text"]) for item in negatives}

    candidates, skipped = promotion_candidates(
        search_match_report,
        existing_exact_terms,
        existing_negative_terms,
    )

    promoted_terms = [item["searchTerm"] for item in candidates]
    newly_negative_terms = [
        item["searchTerm"] for item in candidates if item["normalized"] not in existing_negative_terms
    ]

    add_keywords_result = None
    add_negatives_result = None

    if promoted_terms:
        add_args = [
            "keywords",
            "add",
            "--campaignId",
            str(CATEGORY_CAMPAIGN_ID),
            "--adGroupId",
            str(CATEGORY_EXACT_ADGROUP_ID),
        ]
        for term in promoted_terms:
            add_args.extend(["--text", term])
        add_args.extend(
            [
                "--matchType",
                "EXACT",
                "--bidAmount",
                PROMOTION_BID_GBP,
                "--currency",
                "GBP",
            ]
        )
        add_keywords_result = run_oe_ads(*add_args)

        if newly_negative_terms:
            negative_args = [
                "negatives",
                "add",
                "--campaignId",
                str(DISCOVERY_CAMPAIGN_ID),
            ]
            for term in newly_negative_terms:
                negative_args.extend(["--text", term])
            negative_args.extend(["--matchType", "EXACT"])
            add_negatives_result = run_oe_ads(*negative_args)

    summary = {
        "checkpointUtc": checkpoint.isoformat(),
        "range": {
            "startDate": MONITOR_START_DATE,
            "endDate": today,
        },
        "broadKeywords": broad_report,
        "exactKeywords": exact_report,
        "searchMatchTerms": {
            "campaignId": search_match_report.get("campaignId"),
            "adGroupCount": search_match_report.get("adGroupCount"),
            "startDate": search_match_report.get("startDate"),
            "endDate": search_match_report.get("endDate"),
            "totals": {
                **search_match_report.get("totals", {}),
                "cr": search_match_report.get("totals", {}).get("installRate", 0),
            },
            "rows": [
                {
                    **row,
                    "cr": row.get("installRate", 0),
                }
                for row in search_match_report.get("rows", [])
            ],
        },
        "promotion": {
            "criteria": {
                "minWords": 3,
                "highIntent": list(HIGH_INTENT_MARKERS),
                "maxSpend": 2.25,
                "maxCpt": 0.75,
                "minTtr": 0.04,
                "installRule": "installs >= 1 OR (taps >= 2 AND cr >= 0.25)",
            },
            "candidates": candidates,
            "skipped": skipped,
            "promotedTerms": promoted_terms,
            "negativesAddedTerms": newly_negative_terms,
            "keywordAddResult": add_keywords_result,
            "negativeAddResult": add_negatives_result,
        },
    }

    write_json(run_dir / "broad_keywords.json", broad_report)
    write_json(run_dir / "exact_keywords.json", exact_report)
    write_json(run_dir / "search_match_terms.json", search_match_report)
    write_json(run_dir / "summary.json", summary)
    write_json(OUTPUT_ROOT / "latest-summary.json", summary)

    history_path = OUTPUT_ROOT / "history.ndjson"
    history_path.parent.mkdir(parents=True, exist_ok=True)
    with history_path.open("a", encoding="utf-8") as handle:
        handle.write(json.dumps(summary, sort_keys=True) + "\n")

    print(f"[lifeuk-monitor] checkpoint={checkpoint_slug}")
    print(
        "[lifeuk-monitor] broad totals "
        f"impressions={broad_report['totals'].get('impressions', 0)} "
        f"spend={broad_report['totals'].get('spend', 0)} "
        f"ttr={broad_report['totals'].get('ttr', 0)} "
        f"cr={broad_report['totals'].get('cr', 0)}"
    )
    print(
        "[lifeuk-monitor] exact totals "
        f"impressions={exact_report['totals'].get('impressions', 0)} "
        f"spend={exact_report['totals'].get('spend', 0)} "
        f"ttr={exact_report['totals'].get('ttr', 0)} "
        f"cr={exact_report['totals'].get('cr', 0)}"
    )
    print(
        "[lifeuk-monitor] search match totals "
        f"impressions={search_match_report.get('totals', {}).get('impressions', 0)} "
        f"spend={search_match_report.get('totals', {}).get('spend', 0)} "
        f"ttr={search_match_report.get('totals', {}).get('ttr', 0)} "
        f"cr={search_match_report.get('totals', {}).get('installRate', 0)}"
    )
    if promoted_terms:
        print(f"[lifeuk-monitor] promoted exact terms: {', '.join(promoted_terms)}")
    else:
        print("[lifeuk-monitor] promoted exact terms: none")

    return 0


if __name__ == "__main__":
    raise SystemExit(main())
