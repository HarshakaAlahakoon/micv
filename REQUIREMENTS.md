Using any language/platform you feel comfortable in, create a project that can:

Be run from the command line
Makes a request to 'https://au.mitimes.com/careers/apply/secret'
Post to 'https://au.mitimes.com/careers/apply'
Method = "POST"

Include 'Authorization' header with value from 2

JSON formatted body with the following top level fields:
- name
- email
- job_title
- *final_attempt: Optional - On your final attempt submit with value of `true`.
- **extra_information: **Optional - Add any additional fields to represent you in a JSON object, including but not necessarily limited to your relevant personal attributes, experience (whether on the job, or not), and why we should hire you.