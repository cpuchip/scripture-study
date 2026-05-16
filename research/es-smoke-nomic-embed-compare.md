# nomic-embed-text v1.5 vs v2: Practical Differences

**Binding question:** What are the key practical differences between the nomic-embed-text v1.5 and v2 embedding models — dimensions, context length, and intended use?

## Headlines

**1. Context length dropped from 8,192 tokens to 512 tokens.**
This is the single most consequential difference for anyone building RAG or document-search pipelines. The v1.5 model card lists a **8,192-token** sequence length, and the official docs confirm it natively supports scaling past 2,048 tokens with the right RoPE parameters [nomic-embed-text-v1.5 Hugging Face Model Card](https://huggingface.co/nomic-ai/nomic-embed-text-v1.5). By contrast, v2 has a hard **512-token maximum sequence length** [nomic-embed-text-v2-moe Hugging Face Model Card](https://huggingface.co/nomic-ai/nomic-embed-text-v2-moe). If your workload involves embedding long documents, code files, or transcripts, v1.5 is the only viable option of the two.

**2. v2 trades monolingual simplicity for multilingual reach via MoE.**
v1.5 is a 137M-parameter dense model. v2 is a **475M-parameter Mixture-of-Experts (MoE) model with 305M active parameters** (8 experts, top-2 routing) [nomic-embed-text-v2-moe Hugging Face Model Card](https://huggingface.co/nomic-ai/nomic-embed-text-v2-moe). The vendor blog frames v1.5 as “the most popular open source embedder on Hugging Face” — 35M+ downloads, >100 qps on an M2 MacBook, and easy to scale to massive text collections — while positioning v2 as a multilingual embedder that “outperforms other general purpose embedders of its size on the multilingual MMTEB benchmark” and is “ideal for retrieve-rerank workflows” [The Nomic Embedding Ecosystem](https://homepage-tau-sand.vercel.app/blog/posts/embed-ecosystem).

**3. Both support Matryoshka truncation, but v1.5 offers a wider range.**
Both models output 768-dimensional embeddings and support Matryoshka truncation. v1.5 provides validated performance down to 64 dimensions (MTEB 56.10 at dim 64) [nomic-embed-text-v1.5 Hugging Face Model Card](https://huggingface.co/nomic-ai/nomic-embed-text-v1.5). v2 officially advertises flexible dimension from 768 down to 256, with “3x reductions in storage cost with minimal performance degradations” [nomic-embed-text-v2-moe Hugging Face Model Card](https://huggingface.co/nomic-ai/nomic-embed-text-v2-moe). The v2 model card does not list scores below 256, so sub-256 truncation is unverified by the vendor.

**4. v2 carries more deployment friction.**
v2 requires `trust_remote_code=True`, the `einops` dependency, and a custom prefix format (`search_document:` / `search_query:`) — the same prefixes v1.5 uses, but with additional architectural baggage [nomic-embed-text-v2-moe Hugging Face Model Card](https://huggingface.co/nomic-ai/nomic-embed-text-v2-moe). Independent practitioner Simon Willison confirmed the real-world setup steps and noted the model weighs roughly **1.9 GB** [Nomic Embed Text V2 — Simon Willison’s Blog](https://simonwillison.net/2025/Feb/12/nomic-embed-text-v2/). The v2 model card also warns that “resource requirements may be higher than traditional dense models due to MoE architecture,” even though only a subset of parameters is active during inference [nomic-embed-text-v2-moe Hugging Face Model Card](https://huggingface.co/nomic-ai/nomic-embed-text-v2-moe).

## Notable

- **MTEB at full dimension is nearly identical.** v1.5 scores 62.28 on MTEB at 768 dimensions [Text Embedding — Nomic Platform Documentation](https://docs.nomic.ai/atlas/embeddings-and-retrieval/text-embedding). The v2 model card does not quote a direct MTEB average; it emphasizes multilingual MIRACL (65.80) and BEIR (52.86) instead. For pure English retrieval, v1.5 may still be the better benchmarked choice.
- **v1.5 supports binary storage.** The vendor blog highlights a “binary retrieve-rerank” workflow that cuts vector storage costs by up to 100× with “virtually no loss in downstream performance” [The Nomic Embedding Ecosystem](https://homepage-tau-sand.vercel.app/blog/posts/embed-ecosystem). This is not mentioned for v2.
- **Task prefixes are mandatory for both.** Both models need `search_document:` and `search_query:` prefixes. v1.5 will stop requiring `trust_remote_code=True` starting with `transformers >= 5.5.0` and `sentence-transformers >= 5.3.0` for the text-only series; v2 currently still requires it [nomic-embed-text-v1.5 Hugging Face Model Card](https://huggingface.co/nomic-ai/nomic-embed-text-v1.5).

## Skeptical Takes

- **The 512-token limit is a hard ceiling.** If you are embedding anything longer than a couple of paragraphs, v2 forces chunking. The vendor acknowledges this by recommending v2 for retrieve-rerank (surface candidates cheaply, then reorder with something larger), not for end-to-end long-document embedding.
- **MoE efficiency claims need real-world validation.** The model card admits v2’s “resource requirements may be higher than traditional dense models due to MoE architecture.” Fewer active parameters does not automatically mean faster inference on all hardware, especially if the routing overhead or memory bandwidth becomes the bottleneck. Independent latency benchmarks on CPU vs. GPU are scarce.
- **Benchmarks are self-reported.** The MMTEB and MIRACL comparisons in the ecosystem blog are vendor-authored. The numbers are plausible, but they should be treated as marketing until independently reproduced on the public leaderboard.

## Open Questions

1. **How does v2 perform on monolingual English MTEB at 768-dim?** The sources emphasize multilingual MMTEB and MIRACL, but do not state whether v2 beats v1.5 on the standard English MTEB retrieval tasks that v1.5 was optimized for.
2. **What is the real latency difference between v1.5 dense and v2 MoE on consumer CPUs?** v1.5 is marketed at >100 qps on an M2 MacBook. No comparable on-device figure is given for v2.
3. **Does v2 Matryoshka work below 256 dimensions?** v1.5 is validated down to 64. v2 only advertises 768→256. Whether truncation to 128 or 64 is viable with v2 is unanswered by the vendor documentation.

---

**Sources used:** Vendor model cards (v1.5, v2) [1][2], Nomic platform docs [3], vendor ecosystem blog [4], independent practitioner review by Simon Willison [5]. All claims are sourced to one of these five documents; synthesis is explicitly marked where applicable.