### When to build Agent?
- Workflows that previously resisted automation or were impossible to automate
- Systems that have become unwieldy due to extensive and intricate rulesets, making updates costly or error-prone,   for example performing vendor security reviews.
- When your system relies on unstructured data that are virtually impossible to process using just programming

### Agent design foundations
1. Model - LLM powering the agent
2. Tools - external functions that the model can use
3. Instructions - explicit guidelines that define how the model behaves 

Example:
```python
weather_agent = Agent(
	name= "Weather agent",
	instructions="You are a helpful agent who can talk to users about the weather",
    tools=[get_weather],
)
```

#### Selecting a model
Different models have different strengths and tradeoffs, so need to make some tests and decide what's best for your use case

Principles for choosing the right model:
- Set up evals and performance baseline
- Focus on meeting accuracy target with the best model available
- Optimize for cost and latency

[OpenAI guide on selecting models](https://platform.openai.com/docs/guides/model-selection)

### Defining tools
Tools extend your agent’s capabilities by using APIs from underlying applications or systems. For legacy systems without APIs, agents can rely on computer-use models to interact directly with those applications and systems through web and application UIs—just as a human would.

Each tool should have a standardized definition, enabling flexible, many-to-many relationships between tools and agents. Well-documented, thoroughly tested, and reusable tools improve discoverability, simplify version management, and prevent redundant definitions.

Overall agents need 3 types of tools:
- Data tools - allow agents to retrieve data to process it 
- Action - allow agents to perform some actions, interact with databases, external API's etc.
- Orchestration (**for complex multi-agent structures**) - allow agent to orchestrate the work of other agents or hand off tasks

For example, here’s how you would equip the agent defined above with a series of tools when using the Agents SDK:
![[Pasted image 20250509162856.png]]

### Configuring instructions
Instructions are essential for LLM to perform the operations to need in the way you need them too

Best practices for agent instructions:
- Use existing documents or scripts to provide LLM's with detailed instructions
- Prompt agents to break down tasks
- Define clear actions: make sure every step in your routine corresponds with a specific action or output 
- Capture edge cases

You can use advanced models, like o1 or o3-mini, to automatically generate instructions from
existing documents.

### Orchestration
Single-agent system can handle many tasks with each new tool expanding it's capabilities.

Every orchestration approach needs a concept of a 'run', which is typically a loop that lets agents operate until an exit condition is reached (tool calls, certain strucured outputs, errors etc.)

For example, OpenAI Agents SDK there's Runner.run() method which loops until:
- a final output tool is invoked, which defines desired structured output
- the model returns a response without any tool calls

**When making a complex agent system, it is good to use prompt templates to abstract some use cases**

Practical guidelines for splitting agents include:
- Complex logic: when prompts contain many conditional statements and templates get too difficult 
- Tool overload: model has a lot of similar tools. **If there's 15 tools and they all are well-defined, it's okay**

#### Multi-agent systems
While multi-agent systems can be designed in a lot of different ways, there's two good patterns for this:
- Manager (agents as tools): a central "manager" agent coordinates multiple agents via tool calls
- Decentralized (agents handing off to agents): multiple agents operate as peers, hadning off tasks to one another based on their specializations

Multi-agents systems can be represented as graphs, where each edge is:
- a tool call (for manager pattern)
- a hand off (for decentralized pattern)

##### Manager pattern
Good for when you want a single agent to run a workflow and have access to the user (like a support worker)
![[Pasted image 20250509171133.png]]

Example:
![[Pasted image 20250509171220.png]]
##### Decentralized pattern
This is optimal when you don’t need a single agent maintaining central control or synthesis—instead allowing each agent to take over execution and interact with the user as needed.
![[Pasted image 20250509171407.png]]

![[Pasted image 20250509171548.png]]

This pattern uses a handoffs provided in OpenAI agent SDK
### Guardrails
Well-designed guardrails help you manage data privacy risks (for example, preventing system prompt leaks) or reputational risks (for example, enforcing brand aligned model behavior).   You can set up guardrails that address risks you’ve already identified for your use case and layer   in additional ones as you uncover new vulnerabilities. 

**Guardrails are a critical component of any LLM-based deployment, but should be coupled with robust authentication and authorization protocols, strict access controls, and standard software security measures.**

Think of guardrails as a layered defense mechanism: one can provide good protection, using multiple, specialized guardrails together create more resilient agents.
![[Pasted image 20250509171921.png]]
##### Types of guardrails
- Relevance classifier: ensures agent responses stay within intended scope by flagging off-topic queries
- Safety classifier: detects unsafe inputs (jailbreaks or prompt injections)
- PII filter - prevents unnecessary exposure of personally identifiable information (PII) by vetting model output for any potential PII
- Moderation - Flags harmful or inappropriate inputs (hate speech, harassment, violence) 
- Tool safeguards - assess the risk of each tool available to your agent by assigning a rating based on factors like read-only vs write access, reversibility, required accounts permissions and financial impact. Use these risk ratings to trigger automated actions, such as pausing for guardrail checks before executing high-risk functions or escalating to a human if needed.
- Rules-based protection - simple deterministic measures (blocklists, input length limits, regex filters)
- Output validation - Ensures responses align with brand values via prompt engineering and content checks, preventing outputs that   could harm your brand’s integrity.
#### Building Guardrails
Things to keep in mind: 
- focus on data privacy and content safety
- add new guardrails based on real-world edge cases and failures you encounter
- optimize for both security and user experience

This also uses OpenAI agents SDK (This is openai guide, ofc)

Guardrails can be implemented as functions or agents that enforce some policies 

#### Plan for human intervention 
Human intervention is a critical safeguard enabling you to improve an agent’s real-world performance without compromising user experience. It’s especially important early in deployment, helping identify failures, uncover edge cases, and establish a robust evaluation cycle. 

Implementing a human intervention mechanism allows the agent to gracefully transfer control when it can’t complete a task. In customer service, this means escalating the issue   to a human agent. For a coding agent, this means handing control back to the userm.

Two primary triggers typically warrant human intervention: 
- **Exceeding failure thresholds**: Set limits on agent retries or actions. If the agent exceeds  these limits (e.g., fails to understand customer intent after multiple attempts), escalate  to human intervention. 
- **High-risk actions**: Actions that are sensitive, irreversible, or have high stakes should  trigger human oversight until confidence in the agent’s reliability grows. Examples  include canceling user orders, authorizing large refunds, or making payments.
