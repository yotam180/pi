For further testing and improvements of this project, pick a random repo from online of a well known but not too big product (feel free to explore github).

Add a task for:
1. Cloning the project ~/projects/(...repo name...)
2. Examining the tools there + considering if we wanted to make that repository use PI:
    - Are we missing features? For each missing feature create a proper task for adding that feature
    - Should we add some builtin? For instance, if a project uses Rust, a nice builtin to have would be like install rust (for the setup)
3. Then create a final task with dependencies on the other tasks - for actually transforming the repo to using PI, testing it thoroughly and ensuring it can be run using pi commands and we can create shortcuts to conveniently work with it. 
    For any quirk you get there, think - should we solve it in the yaml definitions level, or introduce a new feature in PI to help developers overcome that quirk properly?
    Don't forget to use the DEVELOPMENT version here in this repo of pi, so that we can deploy fixes quickly in each of these three steps...

I think if it's the first time you read this, we should probably create in docs/ a playbook on how to do each of these steps (clone repo + examine tools, creating tasks, transforming repo to use pi) and attach the playbook file into the tickets you create so followup agents can follow the instructions...