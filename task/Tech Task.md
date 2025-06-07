# Tech Test

## Test Purposes

This should test:

- AWS platform knowledge
- C# or Golang
- General code quality
- Logical thinking
- DDD

---

## Test

✏️ *This scenario is realistic but not real - details have been fabricated just for the purpose of creating a useful exercise.*

The scenario is that **Xapien** wants to get on top of their billing — we want our customers to pay us what they should, and not to use our system more than they’ve paid for.

Our customers use Xapien to run reports (also called “enquiries”), and our pricing generally works around how many reports we let them run every month.

We generally have customers on one of four plans:

- **Ultimate**: 1000 reports per month
- **Enterprise**: 500 reports per month
- **Basic**: 100 reports per month
- **Lite**: 20 reports per month *(we’re probably going to phase this one out soon!)*

---

We also let people **trial Xapien** for a limited time:
They can have **10 reports** over the course of a couple of weeks, but then they aren’t allowed any more until they’ve signed up for a full plan.

---

## System Requirements

We want to build a system that tracks the usage of each customer and:

1. **Stops them running more than their allotted reports**

2. **Sends us a notification** when a customer is on track to hit their monthly limit — so that our customer success team can reach out and try and sell an “uplift”.
   > This warning notification is probably more of a “nice-to-have”, but at the very least the system needs to send us a notification when the limit is reached.

3. **Allows them to configure per-user limits** —
   Say I’m an organisation with 500 reports a month and 10 users, I might reasonably set a 50 user/month limit to make sure one employee doesn’t use them all.

4. **Allows us to top up** a particular customer’s monthly report credits at our discretion.
   These will expire at the end of the month.
   > We’re definitely going to need this before releasing because otherwise customers will run out of credits but we won’t be able to help them.

---

Assumptions:

- We already have **organisation** and **user identifiers**, and **enquiry (report) ids**.

---

## Required REST API Endpoints

1. **Set Plan**
   - Accepts: Org identifier + Plan type (`Ultimate`, `Enterprise`, `Basic`, `Lite`)
   - Action: Stores plan type against the org

2. **Set Per-User Limit**
   - Accepts: Org identifier + limit number
   - Action: Stores the per-user limit for that org

3. **Use Report Credit**
   - Accepts: Org identifier + enquiry identifier
   - Action:
     - Checks if this enquiry is allowed (credits remaining for org/user)
     - Updates backend state accordingly
     - Returns HTTP success code
     - If not allowed: Returns HTTP error code

4. **Top Up Customer Account**
   - Accepts: Org identifier + number of report credits to add
   - Action: Adds top-up credits, expiring at end of the month

---

## Additional Implementation Notes

- Use **AWS stack** and either **C#** or **Golang**
- "Per month" = **calendar month**
- **Authentication and access management** are **out of scope**
- Notification sending can be simulated — e.g., a method call that does nothing

---

## Expectations

- We would like to see how far you get on this in **4 hours** (but this is not a hard limit).
- If you cannot complete your solution that’s fine, but make some notes on what else you’d do.
- Follow **Agile development** principles — start with core functionality first.

---

## Evaluation Criteria

- Code quality, and efforts taken to ensure code equality
- Use of **DDD** in backend code
- Appropriate use of **cloud technologies**
- Correct implementation of **business logic**
