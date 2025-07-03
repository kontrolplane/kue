<p align="center">
  <h1 align="center">
    <a href="https://kontrolplane.dev">
      <img width="1500" alt="kontrolplane header" src="./assets/kontrolplane-header.svg">
    </a>
  </h1>
</p>

`Kue` is a terminal user interface (tui) application designed for managing aws sqs (simple queue service). It provides an intuitive and efficient way to interact with your sqs queues directly from the terminal. With Kue, you can easily create, delete, and manage messages within your queues, making it an essential tool for engineers who prefer working within a terminal environment.

<p align="center">
  <img width="1500" alt="kue cassette" src="./assets/cassette.gif">
</p>

## views

- `queue`: overview, details<sup>1</sup>, creation<sup>1</sup>, delete
- `message`: details<sup>1</sup>, creation<sup>1</sup>, delete<sup>1</sup>

<sup>1</sup>: work in progress

## keybindings

- `q`, `esc`, `ctrl+c`: quit/return
- `↑`, `k`: up
- `↓`, `j`: down
- `→`, `l`: right
- `←`, `h`: left
- `ctrl + d`: delete queue/message
- `ctrl + n`: create queue/message
- `?`: help
- `a`: toggle advanced (create queue form)
- `enter`: view
- `space`: select
- `/`: filter

## Advanced queue creation options ✨

When creating a queue (`ctrl+n` from the queue overview) press **`a`** to expand the *Advanced settings* section.
Here you can optionally configure:

| Option | Description |
| ------ | ----------- |
| Visibility timeout (seconds) | How long a received message is hidden from other consumers before becoming visible again (0 – 43,200) |
| Message retention period (seconds) | How long SQS retains a message that is not deleted (60 – 1,209,600) |
| DLQ ARN | ARN of a dead-letter queue to attach. Provide together with *DLQ max receive count*. |
| DLQ max receive count | After this many receives the message is moved to the DLQ (defaults to 5). |
| KMS Key ID | Customer managed KMS key ID/ARN for server-side encryption. Leave empty to use the default SQS-managed key. |

Only fields you fill in are sent to AWS – leaving a field blank keeps the AWS default.

## demonstration

`queue overview`
<p align="center">
  <img width="1500" alt="kue queue overview" src="./assets/pages/queue/overview.png">
</p>

`queue delete`
<p align="center">
  <img width="1500" alt="kue queue delete" src="./assets/pages/queue/delete.png">
</p>

## Contributors

[//]: kontrolplane/generate-contributors-list

<a href="https://github.com/levivannoort"><img src="https://avatars.githubusercontent.com/u/73097785?v=4" title="levivannoort" width="50" height="50"></a>

[//]: kontrolplane/generate-contributors-list

</br>

<p align="center">
  <img width="1500" alt="kontrolplane foter" src="./assets/kontrolplane-footer.svg">
</p>
