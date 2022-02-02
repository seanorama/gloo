#!/usr/bin/env python3
import sys
import subprocess

def json_post_curl(data, dest):
    curl_command = ["curl", "-X", "POST", "-H", "'Content-type: application/json'", "--data", data, dest]
    process = subprocess.Popen(curl_command, stdout=subprocess.PIPE)
    output, error = process.communicate()
    print(f"output: {output}")
    print(f"error: {error}")        # intentionally allow errors to not fail the script, since this is just a notification step

def main():
    # simplistic argument parsing
    if len(sys.argv) != 3:
        return print("Incorrect number of arguments provided.  Aborting.")
    
    github_actor  = sys.argv[1]
    github_run_id = sys.argv[2]
    slack_webhook = f"{github_actor.upper()}_WEBHOOK"

    # simplistic payload assembly
    payload = '{"text":"Hey `' + github_actor + '`! I noticed that your <https://github.com/solo-io/gloo/actions/runs/' + github_run_id + '|PR build has failed>"}'
    json_post_curl(payload, slack_webhook)

if __name__ == "__main__":
    main()
