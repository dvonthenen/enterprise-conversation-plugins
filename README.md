# Enterprise Conversation Plugins for a Symbl.ai

The goal of this repository is to provide both:

- an "App Store"-like experience to enabling and extending functionality for your [Symbl.ai Enterprise Application](https://github.com/dvonthenen/enterprise-reference-implementation)
- a place for community members to share and contribute their general purpose conversation plugins, workflows, ideas, etc

## Brief Recap: Symbl.ai Enterprise Application

If you aren't familiar with the [Symbl.ai Enterprise Application](https://github.com/dvonthenen/enterprise-reference-implementation), you can find more information about this Architecture in the repo found here: [https://github.com/dvonthenen/enterprise-reference-implementation](https://github.com/dvonthenen/enterprise-reference-implementation).

In short, this Enterprise Application Architecture provides a reusable, off-the-shelf implementation for Conversation Analytics that provides the following benfits:

- Build applications with a historical conversation context
- Persist conversation insights (data ownership)
- Build scalable conversation applications
- Company’s business rules/logic pushed into backend server microservices
- Dashboards, dashboards, dashboards
- Historical data is air-gapped and can be backed up
- UI isolation. Change all aspects of UI frameworks without changing the code
- Also supports Asynchronous analysis of data

This is a high-level block diagram for what the architecture looks like...

![Enterprise Reference Architecture](https://github.com/dvonthenen/enterprise-reference-implementation/blob/main/docs/images/enterprise-architecture.png?raw=true)

## Conversation Plugins Available

Middleware Plugins Available:

- the [Historical Insights Plugin](https://github.com/dvonthenen/enterprise-conversation-plugins/tree/main/plugins/historical-insights) which triggers an Application Specific Message of the last 5 mentions of a Topic, Tracker, Entity, etc
- the [Statistical Insights Plugin](https://github.com/dvonthenen/enterprise-conversation-plugins/tree/main/plugins/statistical-insights) which provides the number of times a topic, Tracker, Entity, etc has been mentioned in the past 30 mins, 1 hour, 4 hours, 1 day, 2 days, 1 week and 1 month.

### How Do I Launch These Plugins

To try these plugins out, make sure the [Prerequisite Components](https://github.com/dvonthenen/enterprise-reference-implementation#prerequisite-components) are running before proceeding! Assuming you have already cloned the [Enterprise Reference Implementation](https://github.com/dvonthenen/enterprise-reference-implementation), clone the [Enterprise Conversation Plugins](hhttps://github.com/dvonthenen/enterprise-conversation-plugins) repo to your local laptop.

```bash
foo@bar:~$ git clone git@github.com:dvonthenen/enterprise-conversation-plugins.git
foo@bar:~$ cd enterprise-reference-implementation
```

Start the [Symbl Proxy/Dataminer](https://github.com/dvonthenen/enterprise-reference-implementation/tree/main/cmd/symbl-proxy-dataminer) in the console by running the following commands:

In your first console windows, run:
```bash
foo@bar:~$ cd ${REPLACE WITH YOUR ROOT DIR}/enterprise-reference-implementation
foo@bar:~$ cd ./cmd/symbl-proxy-dataminer
foo@bar:~$ go run cmd.go
```

Start the [Historical Insights Plugin](https://github.com/dvonthenen/enterprise-conversation-plugins/tree/main/plugins/historical-insights) and the [Statistical Insights Plugin](https://github.com/dvonthenen/enterprise-conversation-plugins/tree/main/plugins/statistical-insights)  each in their own console window.

In a second console windows, run the `historical-insights` plugin by executing:
```bash
foo@bar:~$ cd ${REPLACE WITH YOUR ROOT DIR}/enterprise-conversation-plugins
foo@bar:~$ cd ./plugins/historical-insights
foo@bar:~$ go run cmd.go
```

In a third console windows, run the `statistical-insights` plugin by executing:
```bash
foo@bar:~$ cd ${REPLACE WITH YOUR ROOT DIR}/enterprise-conversation-plugins
foo@bar:~$ cd ./plugins/statistical-insights
foo@bar:~$ go run cmd.go
```

> **_NOTE:_** If you want to run additional plugins, you would just open a new terminal windows, change directory into the plugin folder, and run the plugin.

Finally, create a fourth console window to start the [Example Simulated Client App](https://github.com/dvonthenen/enterprise-reference-implementation/tree/main/cmd/example-simulated-client-app) by running the following commands:

```bash
foo@bar:~$ cd ${REPLACE WITH YOUR ROOT DIR}/enterprise-reference-implementation
foo@bar:~$ cd ./cmd/example-simulated-client-app
foo@bar:~$ go run cmd.go
```

Then start talking into the microphone. Since both of these plugins are about aggregating conversations over time, close the `example-simulated-client-app` instance after having mentioned some topics, entities, trackers, etc. Then start up another instance of the `example-simulated-client-app` (which can be done in the same console window) and mention the same topics, entities, trackers, etc as in the previous conversation session. You should start to see some historical and statistical data flowing through to the client when those past insights are triggered.

## Contact Information

You can reach out to the Community via:

- [Google Group][google_group] for this Community Meeting and Office Hours
- Find us by using the [Community Calendar][google_calendar]
- Taking a look at the [Community Meeting and Office Hours Agenda Notes][agenda_doc]. Feel free to add any agenda items!
- Don't want to wait? Contact us through our [Community Slack][slack]
- If you want to do it the old fashion way, our email is community\[at\]symbl\[dot\]ai

[google_group]: https://bit.ly/3Cp5c9D
[google_calendar]: https://bit.ly/3jRGEj4
[agenda_doc]: https://bit.ly/3WH4hcO
[slack]: https://join.slack.com/t/symbldotai/shared_invite/zt-4sic2s11-D3x496pll8UHSJ89cm78CA
