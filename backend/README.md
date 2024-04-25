# README for backend service

## Stuff to check
- [ ] Pipeline or CommandQueue? Consider if and how external components may interact with it
- [ ] Pipeline: Connected stages by channels; Passing of processed data
- [ ] CommandQueue: Communication by dispatched events; different type of handlers for events; How to handle async comminication? 