# Minimal Offline Bias Classifier Prototype

**Binding question:** What minimal offline HTML/JS implementation demonstrates AI bias by letting visitors input skewed training data and observe classification failures?

**Project:** —

**Date:** 2026-05-12

---

## The plan

Build a single `index.html` file that runs a canvas-based shape classifier using k-Nearest Neighbors in pure JavaScript. No build step, no CDN, no network fetch. The file opens directly in Chrome via `file://` and runs offline on any laptop.

Visitors see randomly generated shapes varying in color, aspect ratio, and edge jaggedness. The classifier extracts two visual features, width and color hue, and plots every shape on a 2D scatter diagram beside the canvas. Clusters overlap when the data is skewed; the geometry of the failure is visible before any explanation appears.

Seed the initial state with deliberately biased training data: ten red rockets against one blue rocket. The model retrains instantly on every label change. After each update, it auto-tests on a held-out validation set and draws bright red "OOPS!" overlays on misclassifications where the machine guesses color instead of shape. A sidebar captions the failure in one sentence: "You only showed it red rockets, so it thinks color matters more than shape." Visitors then click or drag to add balanced examples, hit Retrain, and watch the misclassification count drop.

Target interaction time is 60–90 seconds: observe the bug, add counterexamples, verify improvement. An inactivity timer resets all state after 90 seconds, fading the screen to a "Touch to start" overlay. A single staffer never needs to refresh the browser between visitors.

This prototype validates the pedagogical "aha" moment — pattern mismatch made visible — not the final exhibit stack. Therefore the code is disposable proof-of-concept. Swapping in TensorFlow.js or Teachable Machine becomes a separate work item once the interaction design is proven.

## Assumptions

- "Minimal" means minimal deployment friction — a single HTML file opened via `file://`, no build step, no network fetches. This breaks if the venue requires HTTPS, bundled assets, or cross-origin isolation.
- The target audience is ages 5 and up, therefore the bias must be visible without reading comprehension beyond single-word labels. This breaks if the concept requires explaining precision, recall, or feature vectors.
- k-NN counts as "real ML" for this pedagogical goal; visitors do not need a neural network to experience algorithmic bias. This breaks if stakeholders reject anything short of deep learning as genuine AI.
- The prototype is disposable validation code, not foundation code for the final exhibit. This breaks if the intent is to evolve this file into production rather than rewrite it after validation.
- Generic shapes suffice for the first test; space-themed branding can be layered later via CSS. This breaks if the brand team requires Rocket-vs-Asteroid visuals for the very first user test.
- A 90-second idle timeout matches the target interaction length. This breaks if observation shows visitors need 30 seconds or five minutes.

## Risks

- Older children may dismiss k-NN as "just matching" rather than AI. Watch for this in user testing; honest labeling ("the same pattern-matching trick big AI uses") may not fully mitigate the reaction.
- Visitors who ignore the seeded skew and add random data will not see a bias lesson. The initial skewed state must be visually compelling enough to draw attention to the color-shape mismatch immediately.
- Pure JavaScript locks out the webcam and image-classifier path that the broader plan envisions for the final exhibit. The prototype validates interaction design but throws away the code; rebuilding with TensorFlow.js is required.
- The 90-second auto-reset may interrupt a genuinely engaged visitor. Watch for frustration; add a "Still here" touch target if testing shows cutoff mid-fix.

## Next steps

Implement the single HTML file: random shape generation, k-NN on width and hue, scatter plot rendering, seeded skewed data, misclassification highlighting, and idle reset. Test it offline on the target laptop in a single walkthrough. If the "aha" moment lands, queue a user-testing work item with the center's audience; if the bias is not obvious within sixty seconds, iterate on the feature space or the seeded data before adding any camera, network, or build-pipeline dependencies.