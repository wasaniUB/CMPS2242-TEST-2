export async function createFeedback(feedback) {
  try {
    const res = await fetch("http://localhost:4000/feedback/create", {
      method: "POST",
      headers: {
        "Content-Type": "application/json"
      },
      body: JSON.stringify(feedback)
    });

    const data = await res.json();

    if (!res.ok) {
      return { success: false, error: data.error };
    }

    return { success: true, data };

  } catch (err) {
    return { success: false, error: err.message };
  }
}