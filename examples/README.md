# Examples

The following is a 2-staged tutorial on setting up a templating tool to supply developers to convert a simple yaml into a full set of Kubernetes resources to deploy their app

## 1: Create K8s ConfigMap from sourcefiles

### Folder layout
```
┳━ app1
┃  ┣━ 0_serviceaccount.yaml
┃  ┣━ 1_deployment.yaml
┃  ┗━ 2_service.yaml
┣━ cronjob1
┃  ┣━ 0_serviceaccount.yaml
┃  ┣━ 1_cronjob.yaml
⠇
```

> Note: Most of the files above are empty in the repo and only used as named placeholders

We're using `tree` to generate the input object to our template
```sh
$ tree . -J -L 2 --noreport >tree.json
```

`tree.json`
```json
[
  {"type":"directory","name":".","contents":[
    {"type":"directory","name":"app1","contents":[
      {"type":"file","name":"0_serviceaccount.yaml"},
      {"type":"file","name":"1_deployment.yaml"},
      {"type":"file","name":"2_service.yaml"}
    ]},
    {"type":"directory","name":"cronjob1","contents":[
      {"type":"file","name":"0_serviceaccount.yaml"},
      {"type":"file","name":"1_cronjob.yaml"}
    ]}
  ]}
]
```

Then using the '`kustomization.tmpl`' file we produce a proper '`kustomization.yaml`':
```sh
$ tmpl8 -i tree.json kustomization.tmpl >kustomization.yaml
``` 

'`kustomization.yaml`'
```yaml
namespace: default

generatorOptions:
  disableNameSuffixHash: true

configMapGenerator:
  - name: app1
    files:
    - 0_serviceaccount.yaml=app1/0_serviceaccount.yaml
    - 1_deployment.yaml=app1/1_deployment.yaml
    - 2_service.yaml=app1/2_service.yaml
  - name: cronjob1
    files:
    - 0_serviceaccount.yaml=cronjob1/0_serviceaccount.yaml
    - 1_cronjob.yaml=cronjob1/1_cronjob.yaml

resources:
```

and this is then easily deployed to Kuberenetes:
```sh
$ kubectl apply -k .
```

## 2: Use the K8s-templates (as developer or in a CI/CD pipeline)

In this scenario each microservice/app/job have it´s own YAML-file that is version-stamped using the build-steps (and the version-tag is also used to tag the docker-image) 

```sh
$ tmpl8 -i alpha.yaml k8s:default/app1
```

Using the template '`k8s:default/app`' we will get all the templates stored in the previous step and use those as templates for generating you Kubernetes objects

## Suggestions

You could store the templates in a shared and protected namespace, allowing developer (and pipelines) only read-only access to the configmaps.

You should probably setup some policy-solution to enforce whatever restrictions you want to use.