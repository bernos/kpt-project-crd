package main

import (
  "os"
  "bytes"
  "text/template"
  "sigs.k8s.io/kustomize/kyaml/fn/framework"
  "sigs.k8s.io/kustomize/kyaml/kio/filters"
  "sigs.k8s.io/kustomize/kyaml/yaml"
)

type Project struct {
  Metadata struct {
    Name string `yaml:"name"`
  } `yaml:"metadata"`

  Spec struct {
    Description string `yaml:"description"`
  } `yaml:"spec"`
}

func main() {
  functionConfig := &Project{}
  resourceList := &framework.ResourceList{
    FunctionConfig: functionConfig,
  }

  cmd := framework.Command(resourceList, func() error {
    buf := &bytes.Buffer{}
    t := template.Must(template.New("app-project").Parse(appProjectTemplate))
    if err := t.Execute(buf, functionConfig); err != nil {
      return err
    }

    s, err := yaml.Parse(buf.String())
    if err != nil {
      return err
    }

    _, err = s.Pipe(yaml.LookupCreate(yaml.SequenceNode, "spec", "sourceRepos"), yaml.Append(&yaml.Node{Value: "foo", Kind: yaml.ScalarNode}))
    if err != nil {
      return err
    }

    d := map[string]interface{}{
      "namespace": "foo",
      "server": "bar",
    }

    destination, err := yaml.FromMap(d)
    if err != nil {
      return err
    }

    _, err = s.Pipe(yaml.LookupCreate(yaml.SequenceNode, "spec", "destinations"), yaml.Append(destination.YNode()))
    if err != nil {
      return err
    }

    resourceList.Items = append(resourceList.Items, s)

    resourceList.Items, err = filters.MergeFilter{}.Filter(resourceList.Items)
    if err != nil {
      return err
    }

    resourceList.Items, err = filters.FormatFilter{}.Filter(resourceList.Items)
    if err != nil {
      return err
    }

    return nil
  })

  if err := cmd.Execute(); err != nil {
    os.Exit(1)
  }
}

var namespaceTemplate = `
apiVersion: v1
kind: Namespace
metadata:
  annotations:
    argocd.argoproj.io/sync-wave: "-2"
  name: {{.Metadata.Name}}
`

var appProjectTemplate = `
apiVersion: argoproj.io/v1alpha1
kind: AppProject
metadata:
  annotations:
    argocd.argoproj.io/sync-wave: "-1"
  finalizers:
    - resources-finalizer.argocd.argoproj.io
  name: {{.Metadata.Name}}
  namespace: argocd
spec:
  description: {{.Spec.Description}}
`
