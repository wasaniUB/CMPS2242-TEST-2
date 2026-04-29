export function render(response) {
  const form = document.getElementById("feedbackForm")

  if (response.success) {
     alert("Feedback submitted successfully!")
    form.reset();
  } else {
    alert(`Error in submission: ${response.error}`)
  }
}