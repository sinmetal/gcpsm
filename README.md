# gcpsm
GCP Secret Management Demo Tool

## How it works

### Storage

Registered values are encrypted using Cloud KMS in Cloud Datastor and saved.

For example, when key = sample, value = hoge is registered.

![Secret Register](https://user-images.githubusercontent.com/446022/38799532-abc44f70-419f-11e8-9646-207754cd0ac0.png)

In Datastore, the value encrypted with Cloud KMS is saved.

![Datastore](https://user-images.githubusercontent.com/446022/38799579-c9acf384-419f-11e8-936a-7220a16670e6.png)


### Authentication

* [Google Cloud Identity-Aware Proxy](https://cloud.google.com/iap/)
* [Google App Engine Firewall](https://cloud.google.com/appengine/docs/standard/go/creating-firewalls)

With IAP, you can control with Google Account and GCP Service Account. When using Service Account in CI, it is good to use [iap_curl](https://github.com/b4b4r07/iap_curl) .
If you want to control with IP Addr, use App Engine Firewall.
IAP and Firewall can be used at the same time.

### Monitoring

Save the App Engine log to BigQuery and check it with DataStudio.

![DataStudio](https://user-images.githubusercontent.com/446022/38800390-1de2fda2-41a2-11e8-8ec3-4cb9b52bd5d3.png)
