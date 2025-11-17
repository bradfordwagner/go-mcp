package argo

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("List Applications", func() {
	var testApps []v1alpha1.Application

	BeforeEach(func() {
		testApps = []v1alpha1.Application{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "app1",
					Namespace: "argocd",
				},
				Spec: v1alpha1.ApplicationSpec{
					Project: "project-a",
					Destination: v1alpha1.ApplicationDestination{
						Server: "https://kubernetes.default.svc",
					},
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "app2",
					Namespace: "argocd",
				},
				Spec: v1alpha1.ApplicationSpec{
					Project: "project-a",
					Destination: v1alpha1.ApplicationDestination{
						Server: "https://remote-cluster.example.com",
					},
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "app3",
					Namespace: "argocd",
				},
				Spec: v1alpha1.ApplicationSpec{
					Project: "project-b",
					Destination: v1alpha1.ApplicationDestination{
						Server: "https://kubernetes.default.svc",
					},
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "app4",
					Namespace: "other-namespace",
				},
				Spec: v1alpha1.ApplicationSpec{
					Project: "project-b",
					Destination: v1alpha1.ApplicationDestination{
						Server: "https://remote-cluster.example.com",
					},
				},
			},
		}
	})

	Describe("filterApplications", func() {
		Context("with no filters", func() {
			It("should return all applications", func() {
				input := ListApplicationsInput{}
				result := filterApplications(testApps, input)
				Expect(result).To(HaveLen(len(testApps)))
			})
		})

		DescribeTable("single filter tests",
			func(input ListApplicationsInput, expectedCount int, validateFunc func([]v1alpha1.Application)) {
				result := filterApplications(testApps, input)
				Expect(result).To(HaveLen(expectedCount))
				if validateFunc != nil {
					validateFunc(result)
				}
			},
			Entry("filter by project-a",
				ListApplicationsInput{Project: "project-a"},
				2,
				func(apps []v1alpha1.Application) {
					for _, app := range apps {
						Expect(app.Spec.Project).To(Equal("project-a"))
					}
				},
			),
			Entry("filter by namespace argocd",
				ListApplicationsInput{Namespace: "argocd"},
				3,
				func(apps []v1alpha1.Application) {
					for _, app := range apps {
						Expect(app.Namespace).To(Equal("argocd"))
					}
				},
			),
			Entry("filter by default cluster",
				ListApplicationsInput{Cluster: "https://kubernetes.default.svc"},
				2,
				func(apps []v1alpha1.Application) {
					for _, app := range apps {
						Expect(app.Spec.Destination.Server).To(Equal("https://kubernetes.default.svc"))
					}
				},
			),
			Entry("filter by non-existent project",
				ListApplicationsInput{Project: "non-existent-project"},
				0,
				nil,
			),
		)

		Context("with combined filters", func() {
			It("should match all criteria", func() {
				input := ListApplicationsInput{
					Project:   "project-a",
					Namespace: "argocd",
					Cluster:   "https://kubernetes.default.svc",
				}

				result := filterApplications(testApps, input)
				Expect(result).To(HaveLen(1))
				Expect(result[0].Name).To(Equal("app1"))
				Expect(result[0].Spec.Project).To(Equal("project-a"))
				Expect(result[0].Namespace).To(Equal("argocd"))
				Expect(result[0].Spec.Destination.Server).To(Equal("https://kubernetes.default.svc"))
			})

			It("should return empty when no apps match all filters", func() {
				input := ListApplicationsInput{
					Project:   "project-a",
					Namespace: "other-namespace",
				}

				result := filterApplications(testApps, input)
				Expect(result).To(BeEmpty())
			})
		})

		Context("with empty input", func() {
			It("should handle empty application list", func() {
				emptyApps := []v1alpha1.Application{}
				input := ListApplicationsInput{}

				result := filterApplications(emptyApps, input)
				Expect(result).To(BeEmpty())
			})
		})
	})
})
