# qLog
🗒 plan your day not like a pleb

## Why?

So I suspected I got ADHD by the uncountable amount of projects I keep coming up with.
And all the Ideas were too interesting to not do.
So I realized at this scale I will need a Product Manager.
And I can't afford that cuz I am a student and deathly broke.
So the next best thing in the big 26, Artificial Intelligence.

I tried using a buncha tools. Gemini, Claude, ChatGPT.
And I noticed eventually I just end up describing my project to them anyway.
While describing it I get new ideas, and then I go write that down in a document.
And the thing I learnt is 5 hours of engineering is better than 5 mins of writing.

Another issue was I have to re-explain everything to the AI again cuz it forgets,
even within the same chat. How do you solve that?
you make another AI write the project into a file.
You pass that file as context when you wanna talk to your chatbot again.
The AI also writes detailed docs on everything you discussed,
using a template from a config file (a text file I inject).
All of it is stored in markdown files so you can just cat it if you want.

We can then have the AI write an agenda for tomorrow that gets displayed.
So you come back the next day and go
"Huh lemme check what I need to do today", just run `qlog agenda` and you get your clean file.

Also making it as a TUI cuz I don't wanna do UI.
TUI is cooler, you need info and its right there in your terminal.
I don't wanna leave my cozy terminal and I use neovim anyway so most of my workflow is in the term.

Why Go? I have an issue of choosing a new language for every project I start,
so I learn the language along the way too.

## HOW?

I decided to use:
- BubbleTea for the TUI
- Bubbles for its components
- Lipgloss to make it look cute
- Ollama for the AI

For my setup I use Qwen 3 for the main chatbot cuz it holds a conversation well on my potato,
and Gemma3 to summarize cuz it's insanely fast on my laptop.
But you can choose your own AIs.

## CHECKLIST

- [x] Complete the chat screen
- [ ] Complete the saving to file part
- [ ] Complete the module to read saved files
- [x] Write a markdown file parser
- [ ] Complete the workflow to write to agenda file
- [x] Display agenda using the markdown parser
- [x] Complete the main menu
- [ ] Add fzf search for project files


> [!NOTE]
> WIP, up to usable standard
