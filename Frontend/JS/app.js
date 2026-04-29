import { createFeedback } from "./dataService.js";
import { render } from "./ui.js";

const feedbackForm = document.getElementById("feedbackForm");

feedbackForm.addEventListener("submit", async (e) => {
  e.preventDefault();

  try {
    const feedback = {
      name: document.getElementById("name").value,
      email: document.getElementById("email").value,
      subject: document.getElementById("subject").value,
      message: document.getElementById("message").value
    };

    if (!feedback.name || !feedback.email || !feedback.subject || !feedback.message) {
      render({ success: false, error: "All fields are required" });
      return;
    }

    const result = await createFeedback(feedback);

    render(result);

  } catch (err) {
    console.error("Submission error:", err)
  }
});
