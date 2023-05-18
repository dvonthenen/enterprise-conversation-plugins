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

**Realtime Middleware Plugins Available:**

- the [Historical Plugin](https://github.com/dvonthenen/enterprise-conversation-plugins/tree/main/plugins/realtime/historical) which triggers an Application Specific Message of the last 5 mentions of a Topic, Tracker, Entity, etc
- the [Statistical Plugin](https://github.com/dvonthenen/enterprise-conversation-plugins/tree/main/plugins/realtime/statistical) which provides the number of times a topic, Tracker, Entity, etc has been mentioned in the past 30 mins, 1 hour, 4 hours, 1 day, 2 days, 1 week and 1 month.

**Asynchronous Middleware Plugins Available:**

- the [Email Plugin](https://github.com/dvonthenen/enterprise-conversation-plugins/tree/main/plugins/asynchronous/email) sends an email when a configured Topic, Tracker or Entity is encountered
- the [Webhook Plugin](https://github.com/dvonthenen/enterprise-conversation-plugins/tree/main/plugins/asynchronous/webhook) sends a JSON of the entire conversation to a specified URI when a configured Topic, Tracker or Entity is encountered

### How Do I Launch These Plugins

Please visit the [Enterprise Reference Implementation](https://github.com/dvonthenen/enterprise-reference-implementation) repo for more information. There are 3 main configurations for the implementation contained in that repo and you can find those configurations below.

**Realtime Conversation Processing**
To deploy this configuration, follow this setup guide: [https://github.com/dvonthenen/enterprise-reference-implementation/tree/main/docs/realtime-setup.md](https://github.com/dvonthenen/enterprise-reference-implementation/tree/main/docs/realtime-setup.md).

**Asynchronous Conversation Processing**
To deploy this configuration, follow this setup guide: [https://github.com/dvonthenen/enterprise-reference-implementation/tree/main/docs/asynchronous-setup.md](https://github.com/dvonthenen/enterprise-reference-implementation/tree/main/docs/asynchronous-setup.md).

**Realtime and Asynchronous Conversation Process**
To deploy this configuration, follow each setup guide above. The setup is independent of each other and the only share components between these two configurations are the [Neo4J](https://neo4j.com/) Database and [RabbitMQ](https://rabbitmq.com/) Server.

## Contact Information

You can reach out to the Community via:

- [Google Group][google_group] for this Community Meeting and Office Hours
- Find us by using the [Community Calendar][google_calendar]
- Taking a look at the [Community Meeting][community_meeting]. Feel free to add any topics to [agenda doc][agenda_doc]!
- Bring all questions to the [Office Hours][office_hours].
- Don't want to wait? Contact us through our [Community Slack][slack]
- If you want to do it the old fashion way, our email is symblai-community-meeting\[at\]symbl\[dot\]ai

[google_group]: https://bit.ly/3Cp5c9D
[google_calendar]: https://bit.ly/3jRGEj4
[agenda_doc]: https://bit.ly/3WH4hcO
[community_meeting]: bit.ly/3M13vDg
[office_hours]: bit.ly/3LTbELg
[slack]: https://join.slack.com/t/symbldotai/shared_invite/zt-4sic2s11-D3x496pll8UHSJ89cm78CA
